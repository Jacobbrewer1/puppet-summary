//go:build wireinject
// +build wireinject

package main

import (
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/google/wire"
	"github.com/gorilla/mux"
)

func initializeApp() (*app, error) {
	wire.Build(
		wire.Value(logging.Name(appName)),
		logging.NewConfig,
		logging.CommonLogger,
		mux.NewRouter,
		newApp,
	)
	return new(app), nil
}
