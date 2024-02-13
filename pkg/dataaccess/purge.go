package dataaccess

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

func Purge(from time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Purge the data from the database out of the given range.
	affected, err := DB.Purge(ctx, from)
	if err != nil {
		return fmt.Errorf("error purging data: %w", err)
	} else {
		slog.Debug("Data purged from database", slog.Int("affected", affected))
	}

	// Purge the data from Files out of the given range.
	affected, err = Files.Purge(ctx, from)
	if err != nil {
		return fmt.Errorf("error purging data: %w", err)
	} else {
		slog.Debug("Data purged from Files", slog.Int("affected", affected))
	}
	return nil
}

func PurgeAll() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get the date for tomorrow
	tomorrow := time.Now().AddDate(0, 0, 1)

	// Set the time to midnight
	tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

	// Purge the data from the database out of the given range.
	affected, err := DB.Purge(ctx, tomorrow)
	if err != nil {
		return fmt.Errorf("error purging data: %w", err)
	} else {
		slog.Debug("Data purged from database", slog.Int("affected", affected))
	}

	// Purge the data from Files out of the given range.
	affected, err = Files.Purge(ctx, tomorrow)
	if err != nil {
		return fmt.Errorf("error purging data: %w", err)
	} else {
		slog.Debug("Data purged from Files", slog.Int("affected", affected))
	}

	return nil
}
