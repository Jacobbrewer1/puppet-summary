package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"strings"

	svc "github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/dataaccess"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/Jacobbrewer1/puppet-summary/pkg/request"
	"github.com/google/subcommands"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type serveCmd struct {
	// authToken is the token used to authenticate requests to the upload endpoint. If empty, the endpoint is not secure.
	authToken string

	// autoPurge is the number of days to keep data for. If 0 (or not set), data will not be purged.
	autoPurge int

	// dbType is the type of database to use.
	dbType string

	// gcs is the name of the Google Cloud Storage bucket to use. Setting this will enable GCS.
	gcs string
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
	f.StringVar(&s.authToken, "auth-token", "", "The Bearer token used to authenticate requests to the upload endpoint.")
	f.IntVar(&s.autoPurge, "auto-purge", 0, "The number of days to keep data for. If 0 (or not set), data will not be purged.")
	f.StringVar(&s.dbType, "db", dataaccess.DbSqlite.String(), "The type of database to use. Valid values are 'sqlite', 'mysql', and 'mongodb'.")
	f.StringVar(&s.gcs, "gcs", "", "The name of the Google Cloud Storage bucket to use. (Setting this will enable GCS)")
}

func (s *serveCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if err := setupLogging(); err != nil {
		fmt.Println("Error setting up logging:", err)
		return subcommands.ExitFailure
	}
	s.dbType = strings.TrimSpace(s.dbType)
	s.dbType = strings.ToUpper(s.dbType)
	if !dataaccess.DbOpt(s.dbType).Valid() {
		slog.Error("Invalid database option", slog.String("dbType", s.dbType))
		f.Usage()
		return subcommands.ExitUsageError
	}

	r := mux.NewRouter()

	if err := s.generateConfig(ctx); err != nil {
		slog.Error("Error generating configuration", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}
	s.setupRoutes(r)

	slog.Info(
		"Starting application",
		slog.String("dbType", s.dbType),
		slog.String("gcs", s.gcs),
		slog.Int("autoPurge", s.autoPurge),
		slog.String("commit", Commit),
		slog.String("runtime", fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)),
		slog.String("date", Date),
	)

	// Set up the purge routine
	if s.autoPurge != 0 {
		if err := setupPurge(s.autoPurge); err != nil {
			slog.Error("Error setting up purge routine", slog.String(logging.KeyError, err.Error()))
		}
	} else {
		slog.Info("Auto purge not set, data will not be purged")
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// Start the server in a goroutine, so we can listen for the context to be done.
	go func(srv *http.Server) {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			slog.Info("Server closed gracefully")
		} else if err != nil {
			slog.Error("Error serving requests", slog.String(logging.KeyError, err.Error()))
		}
	}(srv)

	<-ctx.Done()
	slog.Info("Shutting down application")
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("Error shutting down application", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (s *serveCmd) generateConfig(ctx context.Context) error {
	err := dataaccess.ConnectDatabase(ctx, s.dbType)
	if err != nil {
		return fmt.Errorf("error connecting to database: %w", err)
	}
	if s.gcs != "" {
		err = dataaccess.ConnectStorage(ctx, dataaccess.StoreTypeGCS, s.gcs)
		if err != nil {
			return fmt.Errorf("error connecting to Files: %w", err)
		}
	} else {
		err = dataaccess.ConnectStorage(ctx, dataaccess.StoreTypeLocal, "")
		if err != nil {
			return fmt.Errorf("error connecting to local storage: %w", err)
		}
	}
	if s.authToken != "" {
		slog.Info("Upload token set, security on upload endpoint is enabled")
		authToken = s.authToken
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

func (s *serveCmd) setupRoutes(r *mux.Router) {
	r.HandleFunc(pathMetrics, promhttp.Handler().ServeHTTP).Methods(http.MethodGet)
	r.HandleFunc(pathHealth, healthHandler()).Methods(http.MethodGet)

	r.NotFoundHandler = request.NotFoundHandler()
	r.MethodNotAllowedHandler = request.MethodNotAllowedHandler()

	r.PathPrefix(pathAssets).Handler(http.StripPrefix(pathAssets, http.FileServer(http.Dir("./assets"))))

	svc.HandlerWithOptions(new(webService), svc.GorillaServerOptions{
		BaseRouter: r,
		BaseURL:    pathApi,
		ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
			encErr := json.NewEncoder(w).Encode(request.NewMessage(fmt.Sprintf("Error handling request: %s", err)))
			if encErr != nil {
				slog.Error("Error encoding response", slog.String(logging.KeyError, encErr.Error()))
			}
		},
	})
}
