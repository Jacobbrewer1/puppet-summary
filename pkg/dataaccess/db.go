package dataaccess

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/vault"
	"github.com/spf13/viper"
)

const EnvDbConnStr = "DB_CONN_STR"

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
	GetRunsByState(ctx context.Context, states ...summary.State) ([]*entities.PuppetRun, error)

	// GetReports returns all PuppetReports from the database for the given fqdn.
	GetReports(ctx context.Context, fqdn string) ([]*entities.PuppetReportSummary, error)

	// GetReport returns the PuppetReport from the database for the given id.
	GetReport(ctx context.Context, id string) (*entities.PuppetReport, error)

	// GetHistory returns the PuppetHistory from the database for the given environment.
	GetHistory(ctx context.Context, environment ...summary.Environment) ([]*entities.PuppetHistory, error)

	// GetEnvironments returns all environments from the database.
	GetEnvironments(ctx context.Context) ([]summary.Environment, error)

	// Purge purges the data from the database out of the given range.
	Purge(ctx context.Context, from time.Time) (int, error)
}

func ConnectDatabase(ctx context.Context, dbType string, v *viper.Viper) (Database, error) {
	dbType = strings.TrimSpace(dbType)
	dbType = strings.ToUpper(dbType)

	opt := DbOpt(dbType)
	if !opt.Valid() {
		panic("Invalid database option")
	}

	switch opt {
	case DbMongo:
		mongo, err := NewMongo(ctx, v)
		if err != nil {
			return nil, fmt.Errorf("connect to mongo: %w", err)
		}
		return mongo, nil
	case DbMySQL:
		mysql, err := NewMySQL(v)
		if err != nil {
			return nil, fmt.Errorf("connect to mysql: %w", err)
		}
		return mysql, nil
	case DbSqlite:
		sqlite, err := NewSQLite()
		if err != nil {
			return nil, fmt.Errorf("connect to sqlite: %w", err)
		}
		return sqlite, nil
	default:
		return nil, fmt.Errorf("invalid database option, %s", dbType)
	}
}

func GenerateConnectionStr(v *viper.Viper, vs vault.Secrets) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=90s&multiStatements=true&parseTime=true",
		vs.Data["username"],
		vs.Data["password"],
		v.GetString("db.host"),
		v.GetString("db.schema"),
	)
}
