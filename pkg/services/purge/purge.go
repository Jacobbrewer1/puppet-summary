package purge

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/robfig/cron/v3"
)

func (s service) SetupPurge(purgeDays int) error {
	c := cron.New(
		cron.WithLocation(time.UTC),
		cron.WithParser(
			cron.NewParser(
				cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
			),
		),
	)

	// Add a new entry to the cron scheduler to purge every day at 03:00.
	if _, err := c.AddFunc("0 3 * * *", func() {
		err := s.PurgeData(purgeDays)
		if err != nil {
			slog.Error("Error purging data", slog.String(logging.KeyError, err.Error()))
		}
	}); err != nil {
		return fmt.Errorf("error adding purge job to cron scheduler: %w", err)
	}

	c.Start()
	slog.Info("Cron scheduler started")
	return nil
}

func (s service) PurgeData(purgeDays int) error {
	slog.Info("Purging data")

	if purgeDays == 0 {
		slog.Warn("Purge days not set, will not purge any data")
		return nil
	}

	// Get the start and end dates for the purge.
	now := time.Now()
	from := now.AddDate(0, 0, -purgeDays)
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	if purgeDays < 0 {
		// If the purgeDays is <= 0, purge all data.
		slog.Info("Purge days set to 0, purging all data")
		from = time.Now().AddDate(0, 0, 1)
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())
	}

	ctx := context.Background()

	dbAffected, err := s.db.Purge(ctx, from)
	if err != nil {
		return fmt.Errorf("error purging data: %w", err)
	}

	slog.Info("Data purged from Database interface", slog.Int("affected", dbAffected))

	filesAffected, err := dataaccess.Files.Purge(ctx, from)
	if err != nil {
		return fmt.Errorf("error purging data: %w", err)
	}
	slog.Info("Data purged from Files interface", slog.Int("affected", filesAffected))
	slog.Info("Purging complete")

	return nil
}
