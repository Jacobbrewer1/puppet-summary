package main

import (
	"flag"
	"fmt"

	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
)

const (
	// appName is the name of the application.
	appName = "summary"
)

var uploadToken = flag.String("upload-token", "", "The Bearer token used to authenticate requests to the upload endpoint.")

func generateConfig() error {
	err := dataaccess.ConnectDatabase()
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	err = dataaccess.ConnectGCS()
	if err != nil {
		return fmt.Errorf("error connecting to GCS: %w", err)
	}
	return nil
}
