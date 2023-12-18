package dataaccess

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
)

var DB Database

type Database interface {
	// Ping pings the database.
	Ping(ctx context.Context) error

	// SaveRun saves a PuppetRun to the database.
	SaveRun(ctx context.Context, run *entities.PuppetReport) error

	// GetRuns returns all PuppetRuns from the database.
	GetRuns(ctx context.Context) ([]*entities.PuppetRun, error)

	// GetRunsByState returns all PuppetRuns from the database that are in the given state.
	GetRunsByState(ctx context.Context, state entities.State) ([]*entities.PuppetRun, error)

	// GetReports returns all PuppetReports from the database for the given fqdn.
	GetReports(ctx context.Context, fqdn string) ([]*entities.PuppetReportSummary, error)

	// GetReport returns the PuppetReport from the database for the given id.
	GetReport(ctx context.Context, id string) (*entities.PuppetReport, error)

	// GetHistory returns the PuppetHistory from the database for the given environment.
	GetHistory(ctx context.Context, environment entities.Environment) ([]*entities.PuppetHistory, error)

	// GetEnvironments returns all environments from the database.
	GetEnvironments(ctx context.Context) ([]entities.Environment, error)
}

func ConnectDatabase() error {
	flag.Parse()

	*dbFlag = strings.TrimSpace(*dbFlag)
	*dbFlag = strings.ToUpper(*dbFlag)

	opt := dbOpt(*dbFlag)
	if !opt.Valid() {
		panic("Invalid database option")
	}

	switch dbOpt(*dbFlag) {
	case dbMongo:
		connectMongoDB()
	case dbMySQL:
		connectMysql()
	case dbSqlite:
		connectSQLite()
	default:
		// This should never happen as we check for validity in init(). We also have a default of SQLite s
		return fmt.Errorf("invalid database option, %s", *dbFlag)
	}
	return nil
}
