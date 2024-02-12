package dataaccess

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
)

const envDbConnStr = "DB_CONN_STR"

var DB Database

type Database interface {
	// Ping pings the database.
	Ping(ctx context.Context) error

	// Close closes the database connection.
	Close(ctx context.Context) error

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

	// Purge purges the data from the database out of the given range.
	Purge(ctx context.Context, from time.Time) (int, error)
}

func ConnectDatabase(ctx context.Context, dbType string) error {
	dbType = strings.TrimSpace(dbType)
	dbType = strings.ToUpper(dbType)

	opt := DbOpt(dbType)
	if !opt.Valid() {
		panic("Invalid database option")
	}

	switch opt {
	case DbMongo:
		connectMongoDB(ctx)
	case DbMySQL:
		connectMysql()
	case DbSqlite:
		connectSQLite()
	default:
		return fmt.Errorf("invalid database option, %s", dbType)
	}
	return nil
}
