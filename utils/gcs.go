package utils

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"

	"cloud.google.com/go/storage"
)

// Download a file from a GCS bucket to local storage
func DownloadFromGCS(bucketName, objectName string) (*os.File, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	object := bucket.Object(objectName)
	reader, err := object.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("object.NewReader: %v", err)
	}

	tempFile, err := os.CreateTemp("", "download-*.mp4")
	if err != nil {
		return nil, fmt.Errorf("os.CreateTemp: %v", err)
	}

	if _, err := io.Copy(tempFile, reader); err != nil {
		tempFile.Close()
		os.Remove(tempFile.Name())
		return nil, fmt.Errorf("io.Copy: %v", err)
	}

	if _, err := tempFile.Seek(0, 0); err != nil {
		tempFile.Close()
		return nil, fmt.Errorf("tempfile.Seek %v", err)
	}

	return tempFile, nil
}

// Upload a local file to a GCS bucket
func UploadToGCS(ctx context.Context, bucket *storage.BucketHandle, localFilePath, blobPath string) error {
	file, err := os.Open(localFilePath)
	if err != nil {
		return fmt.Errorf("os.Open: %w", err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	wc := bucket.Object(blobPath).NewWriter(ctx)
	defer wc.Close()

	if _, err = io.Copy(wc, reader); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}

	return nil
}
