package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var (
	accountName = os.Getenv("AZURE_STORAGE_ACCOUNT")
	accountKey  = os.Getenv("AZURE_STORAGE_KEY")
	blobURL     = fmt.Sprintf("https://%s.blob.core.windows.net", accountName)
)

const (
	containerName = "quotesubmissions"
	questionPrefix = "questions"
	reflectionPrefix = "reflections"
)

func main() {
	initGroq()
	initBlobClient()

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/submit", submitHandler)

	fmt.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	question, err := getOrGenerateQuestion()
	if err != nil {
		http.Error(w, "Error fetching question: "+err.Error(), http.StatusInternalServerError)
		return
	}

	reflections, err := getTodayReflections()
	if err != nil {
		http.Error(w, "Error fetching reflections: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, struct {
		Question    string
		Reflections []string
	}{
		Question:    question,
		Reflections: reflections,
	})
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	reflection := r.FormValue("reflection")
	if reflection == "" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	date := time.Now().Format("2006-01-02")
	blobName := fmt.Sprintf("reflections/%s.json", date)

	var data map[string]string
	found, _ := blobExists("quotesubmissions", blobName)
	if found {
		rawData, _ := readJSONBlob("quotesubmissions", blobName)
		if mapData, ok := rawData.(map[string]interface{}); ok {
			data = make(map[string]string)
			for k, v := range mapData {
				if str, ok := v.(string); ok {
					data[k] = str
				}
			}
		} else {
			data = make(map[string]string)
		}
	} else {
		data = make(map[string]string)
	}	

	// Find next key
	nextKey := fmt.Sprintf("%d", len(data))
	data[nextKey] = reflection

	err := writeJSONBlob("quotesubmissions", blobName, data)
	if err != nil {
		http.Error(w, "Upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func getOrGenerateQuestion() (string, error) {
	date := time.Now().Format("2006-01-02")
	blobName := fmt.Sprintf("%s.json", date)

	// First check if it exists
	found, err := blobExists("questions", blobName)
	if err != nil {
		return "", err
	}
	if found {
		rawData, err := readJSONBlob("questions", blobName)
		if err != nil {
			return "", err
		}
		if mapData, ok := rawData.(map[string]interface{}); ok {
			if question, ok := mapData["question"].(string); ok {
				return question, nil
			}
		}
		return "", fmt.Errorf("invalid question format")
	}

	// Generate via Groq
	question, err := getDailyQuestionGroq()
	if err != nil {
		return "", err
	}

	// Save it
	err = writeJSONBlob("questions", blobName, map[string]string{"question": question})
	if err != nil {
		return "", err
	}

	return question, nil
}

func getTodayReflections() ([]string, error) {
	date := time.Now().Format("2006-01-02")
	blobName := fmt.Sprintf("reflections/%s.json", date)

	found, err := blobExists("quotesubmissions", blobName)
	if err != nil {
		return nil, err
	}
	if !found {
		return []string{}, nil
	}

	data, err := readJSONBlob("quotesubmissions", blobName)
	if err != nil {
		return nil, err
	}

	// Handle both array and map formats
	switch v := data.(type) {
	case []interface{}:
		var reflections []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				reflections = append(reflections, str)
			}
		}
		return reflections, nil
	case map[string]interface{}:
		var reflections []string
		for _, v := range v {
			if str, ok := v.(string); ok {
				reflections = append(reflections, str)
			}
		}
		return reflections, nil
	default:
		return []string{}, nil
	}
}

func getBlobClient() (*azblob.Client, error) {
	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return nil, err
	}
	return azblob.NewClientWithSharedKeyCredential(blobURL, cred, nil)
}
