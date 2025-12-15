package storage

import (
	"context"
	"io"
)

// Storage is an interface for blob storage operations (firmware files)
type Storage interface {
	// Upload uploads a file to storage and returns the public URL
	Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error)

	// Download retrieves a file from storage
	Download(ctx context.Context, key string) (io.ReadCloser, error)

	// Delete removes a file from storage
	Delete(ctx context.Context, key string) error

	// GetURL returns the public URL for a file
	GetURL(ctx context.Context, key string) (string, error)

	// List lists all files with the given prefix
	List(ctx context.Context, prefix string) ([]string, error)

	// Close closes the storage connection
	Close() error
}
