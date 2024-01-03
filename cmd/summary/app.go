package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// The app is the main application.
type app struct {
	// r is the router.
	r *mux.Router

	// srv is the server.
	srv *http.Server
}

func newApp(_ *slog.Logger, r *mux.Router) *app {
	return &app{
		r: r,
	}
}

func (a *app) run() error {
	if err := a.init(); err != nil {
		return fmt.Errorf("error initializing app: %w", err)
	}

	if err := a.srv.ListenAndServe(); err != nil {
		return fmt.Errorf("error running server: %w", err)
	}

	return nil
}

func (a *app) init() error {
	a.r.HandleFunc(pathIndex, middlewareHttp(indexHandler, AuthOptionNone)).Methods(http.MethodGet)
	a.r.HandleFunc(pathIndexEnv, middlewareHttp(indexHandler, AuthOptionNone)).Methods(http.MethodGet)

	uploadAuth := AuthOptionNone
	if *secureUpload {
		uploadAuth = AuthOptionInternal
	}
	a.r.HandleFunc(pathUpload, middlewareHttp(uploadHandler, uploadAuth)).Methods(http.MethodPost)

	a.r.HandleFunc(pathApiState, middlewareHttp(stateHandler, AuthOptionNone)).Methods(http.MethodGet)
	a.r.HandleFunc(pathRadiator, middlewareHttp(radiatorHandler, AuthOptionNone)).Methods(http.MethodGet)
	a.r.HandleFunc(pathSearch, middlewareHttp(searchHandler, AuthOptionNone)).Methods(http.MethodPost)
	a.r.HandleFunc(pathNodeFqdn, middlewareHttp(nodeFqdnHandler, AuthOptionNone)).Methods(http.MethodGet)
	a.r.HandleFunc(pathReportID, middlewareHttp(reportIDHandler, AuthOptionNone)).Methods(http.MethodGet)

	a.r.HandleFunc(pathMetrics, middlewareHttp(promhttp.Handler().ServeHTTP, AuthOptionInternal)).Methods(http.MethodGet)
	a.r.HandleFunc(pathHealth, middlewareHttp(healthHandler(), AuthOptionInternal)).Methods(http.MethodGet)
	a.r.NotFoundHandler = request.NotFoundHandler()
	a.r.MethodNotAllowedHandler = request.MethodNotAllowedHandler()

	a.r.PathPrefix(pathAssets).Handler(http.StripPrefix(pathAssets, http.FileServer(http.Dir("./assets"))))

	a.srv = &http.Server{
		Addr:    ":8080",
		Handler: a.r,
	}

	return nil
}
