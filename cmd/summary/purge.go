package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/robfig/cron"
)

func setupPurge(purgeDays int) error {
	c := cron.New()

	// Add a new entry to the cron scheduler to purge every (autoPurge) days at 03:00.
	if err := c.AddFunc("0 3 * * *", func() { purgeData(purgeDays) }); err != nil {
		return fmt.Errorf("error adding purge job to cron scheduler: %w", err)
	}

	c.Start()
	slog.Debug("Cron scheduler started")
	return nil
}

func purgeData(purgeDays int) {
	slog.Info("Purging data")

	if purgeDays == 0 {
		slog.Warn("Purge days not set, will not purge any data")
		return
	} else if purgeDays < 0 {
		// If the purgeDays is <= 0, purge all data.
		slog.Info("Purge days set to 0, purging all data")
		err := dataaccess.PurgeAll()
		if err != nil {
			slog.Error("Error purging data", slog.String(logging.KeyError, err.Error()))
			return
		}
		return
	}

	// Get the start and end dates for the purge.
	now := time.Now()
	from := now.AddDate(0, 0, -purgeDays)
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	err := dataaccess.Purge(from)
	if err != nil {
		slog.Error("Error purging data", slog.String(logging.KeyError, err.Error()))
		return
	}

	slog.Info("Purging complete")
}
