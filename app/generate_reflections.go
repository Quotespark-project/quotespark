package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"
	"io"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

func generateAndStoreReflections() error {
	// Use Groq to generate reflections
	prompt := "Generate 10 short, diverse anonymous responses to this question as if written by real people. Format as a numbered list."

	reflections, err := generateDummyReflections(prompt)
	if err != nil {
		return fmt.Errorf("error from Groq: %v", err)
	}

	// Convert reflections array to JSON
	reflectionsJSON, _ := json.Marshal(reflections)

	// Prepare blob client
	account := os.Getenv("AZURE_STORAGE_ACCOUNT")
	key := os.Getenv("AZURE_STORAGE_KEY")
	if account == "" || key == "" {
		return fmt.Errorf("Missing storage account info")
	}

	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		return err
	}
	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", account), cred, nil)
	if err != nil {
		return err
	}

	container := "quotesubmissions"
	date := time.Now().Format("2006-01-02")
	blobName := fmt.Sprintf("reflections/%s.json", date)
	blobClient := client.ServiceClient().NewContainerClient(container).NewBlockBlobClient(blobName)

	// Append to existing content if any
	existing := []byte{}
	getResp, err := blobClient.DownloadStream(context.Background(), nil)
	if err == nil {
		existing, _ = io.ReadAll(getResp.Body)
	}

	newData := append(existing, reflectionsJSON...)

	_, err = blobClient.UploadBuffer(context.Background(), newData, nil)
	if err != nil {
		return fmt.Errorf("upload error: %v", err)
	}

	fmt.Println("✅ Dummy reflections uploaded.")
	return nil
}
