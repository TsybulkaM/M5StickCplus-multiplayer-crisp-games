package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/TsybulkaM/M5StickCplus-multiplayer-crisp-games/internal/core"
)

// NewStorage creates a Storage instance based on the environment
// Uses Azure Blob Storage in production, local filesystem in development
func NewStorage(ctx context.Context) (Storage, error) {
	config := core.LoadConfig()

	storageType := config.StorageType
	if storageType == "" {
		// Auto-detect based on environment
		if config.AzureStorageAccount != "" {
			storageType = "azure"
		} else {
			storageType = "local"
		}
	}

	switch storageType {
	case "azure":
		return newAzureStorage(ctx, config)
	case "local":
		return newLocalStorage(ctx, config)
	default:
		return nil, fmt.Errorf("unknown storage type: %s", storageType)
	}
}

func newAzureStorage(ctx context.Context, config *core.Config) (Storage, error) {
	if config.AzureStorageAccount == "" || config.AzureStorageKey == "" {
		return nil, fmt.Errorf("AZURE_STORAGE_ACCOUNT and AZURE_STORAGE_KEY must be set")
	}

	// Go Cloud CDK URL format for Azure Blob Storage
	// Set environment variables for authentication (required by gocloud.dev/blob/azureblob)
	os.Setenv("AZURE_STORAGE_ACCOUNT", config.AzureStorageAccount)
	os.Setenv("AZURE_STORAGE_KEY", config.AzureStorageKey)

	url := fmt.Sprintf("azblob://%s", config.AzureBlobContainer)
	baseURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s", config.AzureStorageAccount, config.AzureBlobContainer)

	return NewCloudStorage(ctx, url, baseURL)
}

func newLocalStorage(ctx context.Context, config *core.Config) (Storage, error) {
	localPath := config.LocalStoragePath

	// Create directory if it doesn't exist
	if err := os.MkdirAll(localPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	url := fmt.Sprintf("file://%s", absPath)
	// For local storage, baseURL is not used - files are served via /api/fota/download
	return NewCloudStorage(ctx, url, "")
}
