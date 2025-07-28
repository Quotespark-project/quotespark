package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

var blobClient *azblob.Client

func initBlobClient() {
	account := os.Getenv("AZURE_STORAGE_ACCOUNT")
	key := os.Getenv("AZURE_STORAGE_KEY")
	if account == "" || key == "" {
		panic("Storage env vars not set")
	}

	cred, err := azblob.NewSharedKeyCredential(account, key)
	if err != nil {
		panic(err)
	}
	client, err := azblob.NewClientWithSharedKeyCredential(fmt.Sprintf("https://%s.blob.core.windows.net/", account), cred, nil)
	if err != nil {
		panic(err)
	}
	blobClient = client
}

func writeJSONBlob(container, blob string, data interface{}) error {
	ctx := context.Background()
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = blobClient.UploadBuffer(ctx, container, blob, jsonBytes, nil)
	return err
}

func readJSONBlob(container, blob string) (interface{}, error) {
	ctx := context.Background()
	resp, err := blobClient.DownloadStream(ctx, container, blob, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var data interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	return data, err
}

func blobExists(container, blob string) (bool, error) {
	ctx := context.Background()
	containerClient := blobClient.ServiceClient().NewContainerClient(container)
	blobClient := containerClient.NewBlockBlobClient(blob)
	
	_, err := blobClient.GetProperties(ctx, nil)
	if err != nil {
		// Check if it's a 404 error (blob not found)
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "BlobNotFound") {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
