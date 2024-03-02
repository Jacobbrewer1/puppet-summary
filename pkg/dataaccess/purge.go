package dataaccess

import (
	"fmt"
	"time"
)

func Purge(from time.Time) error {
	return nil
	//ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	//defer cancel()
	//
	//// Purge the data from the database out of the given range.
	//affected, err := DB.Purge(ctx, from)
	//if err != nil {
	//	return fmt.Errorf("error purging data: %w", err)
	//} else {
	//	slog.Info("Data purged from database", slog.Int("affected", affected))
	//}
	//
	//// Purge the data from Files out of the given range.
	//affected, err = Files.Purge(ctx, from)
	//if err != nil {
	//	return fmt.Errorf("error purging data: %w", err)
	//} else {
	//	slog.Info("Data purged from Files interface", slog.Int("affected", affected))
	//}
	//return nil
}

func PurgeAll() error {
	// Get the date for tomorrow
	tomorrow := time.Now().AddDate(0, 0, 1)

	// Set the time to midnight
	tomorrow = time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 0, 0, 0, 0, tomorrow.Location())

	err := Purge(tomorrow)
	if err != nil {
		return fmt.Errorf("error purging all data: %w", err)
	}

	return nil
}
