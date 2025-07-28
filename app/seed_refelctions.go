package main

import (
	"fmt"
	"time"
)

func seedDummyReflections() error {
	initGroq()
	initBlobClient()

	date := time.Now().Format("2006-01-02")
	reflectionBlob := fmt.Sprintf("reflections/%s.json", date)

	// Check if reflections already exist
	exists, err := blobExists("quotesubmissions", reflectionBlob)
	if err != nil {
		return fmt.Errorf("failed to check blob: %w", err)
	}
	if exists {
		fmt.Println("Reflections already exist. Skipping seed.")
		return nil
	}

	// Get the question
	questionMap, err := readJSONBlob("quotesubmissions", fmt.Sprintf("%s.json", date))
	if err != nil {
		return fmt.Errorf("could not read question: %w", err)
	}
	
	var reflections []string
	if mapData, ok := questionMap.(map[string]interface{}); ok {
		if question, ok := mapData["question"].(string); ok {
			// Generate dummy reflections
			reflections, err = generateDummyReflections(question)
			if err != nil {
				return fmt.Errorf("GPT error: %w", err)
			}
		}
	}

	// Write to Blob
	err = writeJSONBlob("quotesubmissions", reflectionBlob, reflections)
	if err != nil {
		return fmt.Errorf("blob write error: %w", err)
	}

	fmt.Println("✅ Dummy reflections written.")
	return nil
}
