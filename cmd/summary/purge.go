package main

import (
	"fmt"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"log/slog"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/robfig/cron"
)

func setupPurge(purgeDays int) error {
	if purgeDays == 0 {
		slog.Info("Auto purge not set, data will not be purged")
		return nil
	}

	c := cron.New()

	// Add a new entry to the cron scheduler to purge every (autoPurge) days at 03:00.
	if err := c.AddFunc(fmt.Sprintf("0 3 */%d * *", purgeDays), func() { purgeData(purgeDays) }); err != nil {
		return fmt.Errorf("error adding purge job to cron scheduler: %w", err)
	}

	c.Start()
	slog.Debug("Cron scheduler started")
	return nil
}

func purgeData(purgeDays int) {
	slog.Info("Purging data")

	// Get the start and end dates for the purge.
	now := time.Now()
	from := now.AddDate(0, 0, -purgeDays)

	err := dataaccess.Purge(entities.Datetime(from))
	if err != nil {
		slog.Error("Error purging data", slog.String(logging.KeyError, err.Error()))
		return
	}

	slog.Info("Purging complete")
}
