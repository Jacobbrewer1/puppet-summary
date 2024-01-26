package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/google/subcommands"
)

type purgeCmd struct {
	// days is the number of days to purge.
	days int

	// configPath is the path to the config file.
	configPath string

	// dbType is the type of database to connect to.
	dbType string

	// gcs is whether to connect to GCS.
	gcs bool

	// gcsBucket is the name of the GCS bucket to connect to.
	gcsBucket string
}

func (p *purgeCmd) Name() string {
	return "purge"
}

func (p *purgeCmd) Synopsis() string {
	return "Purge old puppet reports"
}

func (p *purgeCmd) Usage() string {
	return `purge:
  Purge old puppet reports.
`
}

func (p *purgeCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&p.days, "days", 0, "The number of days to purge.")
	f.StringVar(&p.configPath, "config", "", "The path to the config file.")
	f.StringVar(&p.dbType, "db-type", "mysql", "The type of database to connect to.")
	f.BoolVar(&p.gcs, "gcs", false, "Whether to connect to GCS.")
	f.StringVar(&p.gcsBucket, "gcs-bucket", "", "The name of the GCS bucket to connect to.")
}

func (p *purgeCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if p.days == 0 {
		f.Usage()
		return subcommands.ExitUsageError
	}

	// Setup logging
	if err := setupLogging(); err != nil {
		slog.Error("Error setting up logging", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	// Generate the config
	if err := p.generateConfig(context.Background()); err != nil {
		slog.Error("Error generating config", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	// Purge the reports
	purgeData(p.days)

	return subcommands.ExitSuccess
}

func (p *purgeCmd) generateConfig(ctx context.Context) error {
	err := dataaccess.ConnectDatabase(ctx, p.dbType)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	if p.gcs {
		dataaccess.GCSEnabled = true
		err = dataaccess.ConnectGCS(p.gcsBucket)
		if err != nil {
			return fmt.Errorf("error connecting to GCS: %w", err)
		}
	}
	return nil
}
