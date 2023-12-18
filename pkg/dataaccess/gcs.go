package dataaccess

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

const (
	envGCSCredentials = "GCS_CREDENTIALS"
	envGCSBucket      = "GCS_BUCKET"
)

var GCS Storage

type Storage interface {
	// SaveFile uploads a file to the storage bucket. This will replace any existing file with the same name.
	SaveFile(ctx context.Context, filePath string, file []byte) error

	// DownloadFile downloads a file from the storage bucket.
	DownloadFile(ctx context.Context, filePath string) ([]byte, error)

	// DeleteFile deletes a file from the storage bucket.
	DeleteFile(ctx context.Context, filePath string) error
}

type storageImpl struct {
	// gcs is the Google Cloud Storage client.
	gcs *storage.Client

	// bucket is the name of the bucket to use.
	bucket string
}

func newStorage(gcs *storage.Client, bucket string) Storage {
	return &storageImpl{
		gcs:    gcs,
		bucket: bucket,
	}
}

func (s *storageImpl) SaveFile(ctx context.Context, filePath string, file []byte) error {
	if !GCSEnabled {
		// GCS is not enabled, so do nothing.
		return nil
	}

	// Connect to the bucket.
	bkt := s.gcs.Bucket(s.bucket)

	// Create a new file in the bucket.
	w := bkt.Object(filePath).NewWriter(ctx)

	// Write the file to the bucket.
	_, err := w.Write(file)
	if err != nil {
		return fmt.Errorf("error writing file to bucket: %w", err)
	}

	// Close the file.
	err = w.Close()
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	return nil
}

func (s *storageImpl) DownloadFile(ctx context.Context, filePath string) ([]byte, error) {
	if !GCSEnabled {
		// GCS is not enabled, so do nothing.
		return nil, nil
	}

	// Connect to the bucket.
	bkt := s.gcs.Bucket(s.bucket)

	// Open the file.
	r, err := bkt.Object(filePath).NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	// Read the file.
	file, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Close the file.
	err = r.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing file: %w", err)
	}

	return file, nil
}

func (s *storageImpl) DeleteFile(ctx context.Context, filePath string) error {
	if !GCSEnabled {
		// GCS is not enabled, so do nothing.
		return nil
	}

	// Connect to the bucket.
	bkt := s.gcs.Bucket(s.bucket)

	// Delete the file.
	err := bkt.Object(filePath).Delete(ctx)
	if err != nil {
		return fmt.Errorf("error deleting file: %w", err)
	}

	return nil
}

func ConnectGCS() error {
	if !*gcsFlag {
		// GCS is not enabled, so do nothing.
		// Set the GCS variable to an empty storage implementation to prevent nil pointer errors.
		slog.Debug("GCS is not enabled, skipping connection")
		GCS = newStorage(nil, "")
		return nil
	}

	GCSEnabled = true

	// Get the service account credentials from the environment variable.
	gcsCredentials := os.Getenv(envGCSCredentials)
	if gcsCredentials == "" {
		return errors.New("no GCS credentials provided")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(gcsCredentials)))
	if err != nil {
		return fmt.Errorf("error connecting to GCS: %w", err)
	}
	cs := client

	// Get the bucket name from the environment variable and validate that it exists.
	gcsBucket := os.Getenv(envGCSBucket)
	if gcsBucket == "" {
		return errors.New("no GCS bucket provided")
	}

	_, err = cs.Bucket(gcsBucket).Attrs(ctx)
	if err != nil {
		return fmt.Errorf("error validating GCS bucket: %w", err)
	}

	GCS = newStorage(cs, gcsBucket)
	slog.Debug("Connected to GCS")
	return nil
}
