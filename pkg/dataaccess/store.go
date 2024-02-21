package dataaccess

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var Files fileHandler

type fileHandler interface {
	// SaveFile uploads a file to storage. This will replace any existing file with the same name.
	SaveFile(ctx context.Context, filePath string, file []byte) error

	// DownloadFile downloads a file from storage.
	DownloadFile(ctx context.Context, filePath string) ([]byte, error)

	// DeleteFile deletes a file from storage.
	DeleteFile(ctx context.Context, filePath string) error

	// Purge purges the data from storage out of the given range.
	Purge(ctx context.Context, from time.Time) (int, error)
}

func ConnectStorage(ctx context.Context, storeType StoreType, bucketName string) error {
	switch storeType {
	case StoreTypeLocal:
		f, err := newLocal()
		if err != nil {
			return fmt.Errorf("error creating local fileHandler: %w", err)
		}
		Files = f
		return nil
	case StoreTypeGCS:
		return connectGCS(ctx, bucketName)
	case StoreTypeS3:
		return errors.New("s3 fileHandler not implemented yet")
	default:
		return fmt.Errorf("invalid fileHandler type: %s", storeType)
	}
}
