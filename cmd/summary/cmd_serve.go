package main

import (
	"context"
	"flag"
	"log/slog"

	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/google/subcommands"
)

type serveCmd struct {
	// uploadToken is the token used to authenticate requests to the upload endpoint. If empty, the endpoint is not secure.
	uploadToken string

	// autoPurge is the number of days to keep data for. If 0 (or not set), data will not be purged.
	autoPurge int

	// dbType is the type of database to use.
	dbType string

	// gcs is whether to use Google Cloud Storage.
	gcs bool
}

func (s *serveCmd) Name() string {
	return "serve"
}

func (s *serveCmd) Synopsis() string {
	return "Start the web server"
}

func (s *serveCmd) Usage() string {
	return `serve:
  Start the web server.
`
}

func (s *serveCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&s.uploadToken, "upload-token", "", "The Bearer token used to authenticate requests to the upload endpoint.")
	f.IntVar(&s.autoPurge, "auto-purge", 0, "The number of days to keep data for. If 0 (or not set), data will not be purged.")
	f.StringVar(&s.dbType, "db", "sqlite", "The type of database to use. Valid values are 'sqlite', 'mysql', and 'mongodb'.")
	f.BoolVar(&s.gcs, "gcs", false, "Whether to use Google Cloud Storage.")
}

func (s *serveCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	a, err := initializeApp()
	if err != nil {
		slog.Error("Error initializing application", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}
	if err := s.generateConfig(); err != nil {
		slog.Error("Error generating configuration", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}
	slog.Debug("Starting application")
	if err := a.run(s.autoPurge); err != nil {
		slog.Error("Error running application", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
