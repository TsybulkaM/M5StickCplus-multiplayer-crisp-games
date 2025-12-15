package storage

import (
	"context"
	"fmt"
	"io"

	"gocloud.dev/blob"
	_ "gocloud.dev/blob/azureblob" // Azure Blob Storage driver
	_ "gocloud.dev/blob/fileblob"  // Local filesystem driver
)

// CloudStorage implements Storage interface using Go Cloud Development Kit
type CloudStorage struct {
	bucket  *blob.Bucket
	baseURL string
}

// NewCloudStorage creates a new CloudStorage instance
// For Azure: url = "azblob://container-name"
// For Local: url = "file:///path/to/directory"
func NewCloudStorage(ctx context.Context, url string, baseURL string) (*CloudStorage, error) {
	bucket, err := blob.OpenBucket(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to open bucket: %w", err)
	}

	return &CloudStorage{
		bucket:  bucket,
		baseURL: baseURL,
	}, nil
}

// Upload uploads a file to storage
func (s *CloudStorage) Upload(ctx context.Context, key string, data io.Reader, contentType string) (string, error) {
	opts := &blob.WriterOptions{
		ContentType: contentType,
	}

	w, err := s.bucket.NewWriter(ctx, key, opts)
	if err != nil {
		return "", fmt.Errorf("failed to create writer: %w", err)
	}

	if _, err := io.Copy(w, data); err != nil {
		w.Close()
		return "", fmt.Errorf("failed to write data: %w", err)
	}

	if err := w.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	url, err := s.GetURL(ctx, key)
	if err != nil {
		return "", err
	}

	return url, nil
}

// Download retrieves a file from storage
func (s *CloudStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	reader, err := s.bucket.NewReader(ctx, key, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create reader: %w", err)
	}

	return reader, nil
}

// Delete removes a file from storage
func (s *CloudStorage) Delete(ctx context.Context, key string) error {
	if err := s.bucket.Delete(ctx, key); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// GetURL returns the public URL for a file
func (s *CloudStorage) GetURL(ctx context.Context, key string) (string, error) {
	if s.baseURL != "" {
		return fmt.Sprintf("%s/%s", s.baseURL, key), nil
	}

	// If no base URL is provided, try to get a signed URL
	// Note: This may not work for all blob storage types
	return fmt.Sprintf("blob://%s", key), nil
}

// List lists all files with the given prefix
func (s *CloudStorage) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	iter := s.bucket.List(&blob.ListOptions{
		Prefix: prefix,
	})

	for {
		obj, err := iter.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		keys = append(keys, obj.Key)
	}

	return keys, nil
}

// Close closes the storage connection
func (s *CloudStorage) Close() error {
	return s.bucket.Close()
}
