package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func generateAndStoreQuestion() error {
	// Step 1: Get question from Groq
	question, err := getDailyQuestionGroq()
	if err != nil {
		return fmt.Errorf("GPT error: %v", err)
	}

	// Step 2: Prepare Azure Blob client
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
	accountKey := os.Getenv("AZURE_STORAGE_KEY")
	if accountName == "" || accountKey == "" {
		return fmt.Errorf("Azure credentials missing")
	}

	cred, err := azblob.NewSharedKeyCredential(accountName, accountKey)
	if err != nil {
		return fmt.Errorf("Credential error: %v", err)
	}

	serviceURL := fmt.Sprintf("https://%s.blob.core.windows.net/", accountName)
	serviceClient, err := azblob.NewClientWithSharedKeyCredential(serviceURL, cred, nil)
	if err != nil {
		return fmt.Errorf("Service client error: %v", err)
	}

	containerClient := serviceClient.ServiceClient().NewContainerClient("questions")

	// Step 3: Upload today's question
	today := time.Now().Format("2006-01-02")
	filename := fmt.Sprintf("%s.json", today)
	blobClient := containerClient.NewBlockBlobClient(filename)

	questionJSON, _ := json.Marshal(map[string]string{"question": question})

	_, err = blobClient.UploadBuffer(context.TODO(), questionJSON, nil)
	if err != nil {
		return fmt.Errorf("Upload error: %v", err)
	}

	fmt.Println("✅ Daily question uploaded:", question)
	return nil
}
