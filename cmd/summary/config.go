package main

import (
	"fmt"
	"log/slog"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
)

const (
	// appName is the name of the application.
	appName = "summary"
)

var (
	uploadToken string
)

func (s *serveCmd) generateConfig() error {
	err := dataaccess.ConnectDatabase(s.dbType)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	if s.gcs {
		dataaccess.GCSEnabled = true
		err = dataaccess.ConnectGCS(s.gcsBucket)
		if err != nil {
			return fmt.Errorf("error connecting to GCS: %w", err)
		}
	}
	if s.uploadToken != "" {
		slog.Info("Upload token set, security on upload endpoint is enabled")
		uploadToken = s.uploadToken
	} else {
		slog.Info("Upload token not set, upload endpoint is not secure")
	}
	if s.autoPurge != 0 {
		slog.Info(fmt.Sprintf("Auto purge set to %d days", s.autoPurge))
	} else {
		slog.Info("Auto purge not set, data will not be purged")
	}
	return nil
}
