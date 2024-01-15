package dataaccess

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
)

func Purge(from entities.Datetime) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Purge the data from the database out of the given range.
	affected, err := DB.Purge(ctx, from)
	if err != nil {
		return fmt.Errorf("error purging data: %w", err)
	} else {
		slog.Debug("Data purged from database", slog.Int("affected", affected))
	}

	// Purge the data from GCS out of the given range.
	affected, err = GCS.Purge(ctx, from)
	if err != nil {
		return fmt.Errorf("error purging data: %w", err)
	} else {
		slog.Debug("Data purged from GCS", slog.Int("affected", affected))
	}
	return nil
}
