package main

import (
	"fmt"

	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
)

const (
	// appName is the name of the application.
	appName = "summary"
)

var (
	authToken string
)

func setupLogging() error {
	lCfg := logging.NewConfig(appName)

	_, err := logging.CommonLogger(lCfg)
	if err != nil {
		return fmt.Errorf("error creating common logger: %w", err)
	}

	return nil
}
