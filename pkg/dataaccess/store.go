package dataaccess

import (
	"context"
	"errors"
	"fmt"

	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
)

var Files Storage

type Storage interface {
	// SaveFile uploads a file to the storage bucket. This will replace any existing file with the same name.
	SaveFile(ctx context.Context, filePath string, file []byte) error

	// DownloadFile downloads a file from the storage bucket.
	DownloadFile(ctx context.Context, filePath string) ([]byte, error)

	// DeleteFile deletes a file from the storage bucket.
	DeleteFile(ctx context.Context, filePath string) error

	// Purge purges the data from the storage bucket out of the given range.
	Purge(ctx context.Context, from entities.Datetime) (int, error)
}

func ConnectStorage(ctx context.Context, storeType StoreType, bucketName string) error {
	switch storeType {
	case StoreTypeLocal:
		f, err := newLocal()
		if err != nil {
			return fmt.Errorf("error creating local storage: %w", err)
		}
		Files = f
		return nil
	case StoreTypeGCS:
		return connectGCS(ctx, bucketName)
	case StoreTypeS3:
		return errors.New("s3 storage not implemented yet")
	default:
		return fmt.Errorf("invalid storage type: %s", storeType)
	}
}
