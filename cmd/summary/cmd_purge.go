package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/google/subcommands"
)

type purgeCmd struct {
	// days is the number of days to purge.
	days int

	// dbType is the type of database to connect to.
	dbType string

	// gcs is whether to connect to Files.
	gcs string
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
	f.IntVar(&p.days, "days", 0, "The number of days to purge. 0 will not purge any data, <0 will purge all data.")
	f.StringVar(&p.dbType, "db", dataaccess.DbSqlite.String(), "The type of database to connect to.")
	f.StringVar(&p.gcs, "gcs", "", "The name of the Google Cloud storage bucket to use. (Setting this will enable GCS)")
}

func (p *purgeCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if p.days == 0 {
		slog.Warn("Days not set, will not purge any data")
		return subcommands.ExitUsageError
	} else if p.days < 0 {
		// Get confirmation that this will purge all data
		fmt.Println("Purging all data")
		fmt.Print("Are you sure? (yes/no): ")
		var confirm string
		if _, err := fmt.Scanln(&confirm); err != nil {
			slog.Error("Error reading input", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}
		if strings.ToLower(confirm) != "yes" {
			fmt.Println("Purge cancelled")
			return subcommands.ExitUsageError
		}
	}

	p.dbType = strings.TrimSpace(p.dbType)
	p.dbType = strings.ToUpper(p.dbType)
	if !dataaccess.DbOpt(p.dbType).Valid() {
		slog.Error("Invalid database option", slog.String("dbType", p.dbType))
		f.Usage()
		return subcommands.ExitUsageError
	}

	// Setup logging
	if err := setupLogging(); err != nil {
		slog.Error("Error setting up logging", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	// Generate the config
	if err := p.generateConfig(ctx); err != nil {
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
	if p.gcs != "" {
		err = dataaccess.ConnectStorage(ctx, dataaccess.StoreTypeGCS, p.gcs)
		if err != nil {
			return fmt.Errorf("error connecting to Files: %w", err)
		}
	} else {
		err = dataaccess.ConnectStorage(ctx, dataaccess.StoreTypeLocal, "")
		if err != nil {
			return fmt.Errorf("error connecting to local storage: %w", err)
		}
	}
	return nil
}
