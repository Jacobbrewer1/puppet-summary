package dataaccess

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/codegen/apis/summary"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/viper"
)

type mysqlImpl struct {
	// client is the database.
	client *Db
}

func (m *mysqlImpl) Reconnect(ctx context.Context, connStr string) error {
	// Create a new database connection.
	newDb, err := sqlx.Open("mysql", connStr)
	if err != nil {
		return fmt.Errorf("error opening mysql: %w", err)
	}

	err = m.client.Reconnect(ctx, newDb)
	if err != nil {
		return fmt.Errorf("error reconnecting: %w", err)
	}

	return nil
}

func (m *mysqlImpl) Close(_ context.Context) error {
	return m.client.Close()
}

func (m *mysqlImpl) Purge(ctx context.Context, from time.Time) (int, error) {
	sqlStmt := `
	DELETE FROM reports
	WHERE executed_at < ?;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("purge"))
	defer t.ObserveDuration()

	stmt, err := m.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return 0, fmt.Errorf("error preparing statement: %w", err)
	}

	res, err := stmt.ExecContext(ctx, from.Format(time.DateTime))
	if err != nil {
		return 0, fmt.Errorf("error executing statement: %w", err)
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting rows affected: %w", err)
	}

	return int(affected), nil
}

func (m *mysqlImpl) GetEnvironments(ctx context.Context) ([]summary.Environment, error) {
	sqlStmt := `
	SELECT DISTINCT environment 
	FROM reports;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_environments"))
	defer t.ObserveDuration()

	stmt, err := m.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			slog.Error("Error closing rows", slog.String(logging.KeyError, err.Error()))
		}
	}(rows)

	envs := make([]summary.Environment, 0)
	for rows.Next() {
		var env summary.Environment
		if err := rows.Scan(&env); err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}
		envs = append(envs, env)
	}

	return envs, nil
}

func (m *mysqlImpl) GetHistory(ctx context.Context, environment ...summary.Environment) ([]*entities.PuppetHistory, error) {
	res := make([]*entities.PuppetHistory, 0)

	limit := 30

	// Check the environments are valid.
	whereClause := ""
	envStrSlice := make([]any, len(environment))
	if len(environment) > 0 {
		for i, env := range environment {
			if !env.IsValid() {
				return nil, fmt.Errorf("invalid environment: %s", env)
			}

			whereClause += "?"
			if i != len(environment)-1 {
				whereClause += ","
			}

			envStrSlice[i] = string(env)
		}
	}

	query := "SELECT DISTINCT DATE(executed_at) FROM reports;"
	if len(environment) > 0 {
		query = "SELECT DISTINCT DATE(executed_at) FROM reports WHERE environment IN (" + whereClause + ");"
	}

	stmt, err := m.client.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	rows, err := stmt.QueryContext(ctx, envStrSlice...)
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}
	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			slog.Error("Error closing statement", slog.String(logging.KeyError, err.Error()))
		}
	}(stmt)
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			slog.Error("Error closing rows", slog.String(logging.KeyError, err.Error()))
		}
	}(rows)

	dates := make([]string, 0)

	for rows.Next() {
		var d string
		err = rows.Scan(&d)
		if err != nil {
			return nil, errors.New("failed to scan SQL")
		}

		// Parse the date.
		dt, err := time.Parse(time.RFC3339, d)
		if err != nil {
			slog.Error("Error parsing date", slog.String(logging.KeyError, err.Error()))
			continue
		}

		dates = append(dates, dt.Format(time.DateOnly))
	}
	if len(dates) < limit {
		limit = len(dates)
	}

	for _, date := range dates[:limit] {
		x := new(entities.PuppetHistory)
		x.Changed = 0
		x.Unchanged = 0
		x.Failed = 0
		x.Date = date

		startTime, err := time.Parse(time.DateOnly, date)
		if err != nil {
			slog.Error("Error parsing date", slog.String(logging.KeyError, err.Error()))
			continue
		}
		endTime := startTime.AddDate(0, 0, 1)

		locQuery := "SELECT DISTINCT state, COUNT('state') FROM reports WHERE executed_at BETWEEN ? AND ?"
		if len(environment) > 0 {
			locQuery += " AND environment IN ("
			for i := range environment {
				locQuery += "?"
				if i != len(environment)-1 {
					locQuery += ","
				}
			}
			locQuery += ")"
		}

		locQuery += " GROUP BY state;"
		stmt, err = m.client.PrepareContext(ctx, locQuery)
		if err != nil {
			return nil, fmt.Errorf("error preparing statement: %w", err)
		}

		locWhere := make([]any, 0)
		locWhere = append(locWhere, startTime.Format(time.DateOnly))
		locWhere = append(locWhere, endTime.Format(time.DateOnly))
		if len(environment) > 0 {
			locWhere = append(locWhere, envStrSlice...)
		}

		rows, err = stmt.QueryContext(ctx, locWhere...)
		if err != nil {
			return nil, fmt.Errorf("error executing statement: %w", err)
		}

		for rows.Next() {
			var state summary.State
			var count int

			err = rows.Scan(&state, &count)
			if err != nil {
				return nil, errors.New("failed to scan SQL")
			}
			if state.IsIn(summary.State_CHANGED) {
				x.Changed += count
			}
			if state.IsIn(summary.State_UNCHANGED) {
				x.Unchanged += count
			}
			if state.IsIn(summary.State_FAILED) {
				x.Failed += count
			}
		}

		if err := stmt.Close(); err != nil {
			slog.Error("Error closing statement", slog.String(logging.KeyError, err.Error()))
		}

		if err := rows.Close(); err != nil {
			slog.Error("Error closing rows", slog.String(logging.KeyError, err.Error()))
		}

		res = append(res, x)
	}

	return res, nil
}

func (m *mysqlImpl) GetReport(ctx context.Context, id string) (*entities.PuppetReport, error) {
	sqlStmt := `
SELECT hash,
       fqdn,
       environment,
       state, 
       executed_at, 
       runtime, 
       failed, 
       changed, 
       total,
       yaml_file 
FROM reports 
WHERE hash = ?;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_report"))
	defer t.ObserveDuration()

	stmt, err := m.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	row := stmt.QueryRowContext(ctx, id)

	report := new(entities.PuppetReport)
	err = row.Scan(&report.ID, &report.Fqdn, &report.Env, &report.State, &report.ExecTime, &report.Runtime,
		&report.Failed, &report.Changed, &report.Total, &report.YamlFile)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf("error scanning rows: %w", err)
	}

	return report, nil
}

func (m *mysqlImpl) GetReports(ctx context.Context, fqdn string) ([]*entities.PuppetReportSummary, error) {
	sqlStmt := `
SELECT hash, 
       fqdn,
       environment,
       state, 
       executed_at, 
       runtime, 
       failed, 
       changed, 
       total,
       yaml_file 
FROM reports 
WHERE fqdn = ? 
ORDER by executed_at DESC;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_reports"))
	defer t.ObserveDuration()

	stmt, err := m.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	rows, err := stmt.QueryContext(ctx, fqdn)
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}
	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			slog.Error("Error closing rows", slog.String(logging.KeyError, err.Error()))
		}
	}(rows)

	reports := make([]*entities.PuppetReportSummary, 0)
	for rows.Next() {
		report := new(entities.PuppetReportSummary)
		if err := rows.Scan(&report.ID, &report.Fqdn, &report.Env, &report.State, &report.ExecTime, &report.Runtime,
			&report.Failed, &report.Changed, &report.Total, &report.YamlFile); err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}

		report.CalculateTimeSince()

		reports = append(reports, report)
	}

	return reports, nil
}

func (m *mysqlImpl) GetRunsByState(ctx context.Context, states ...summary.State) ([]*entities.PuppetRun, error) {
	sqlStmt := `
	SELECT hash,
		   fqdn,
		   state,
		   executed_at,
		   runtime
	FROM reports
	WHERE state IN (?)
	ORDER BY executed_at DESC;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_runs_by_state"))
	defer t.ObserveDuration()

	stmt, err := m.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	// Convert the states to a string csv.
	statesStr := ""
	for i, state := range states {
		statesStr += string(state)
		if i != len(states)-1 {
			statesStr += ","
		}
	}

	rows, err := stmt.QueryContext(ctx, statesStr)
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Error closing rows", slog.String(logging.KeyError, err.Error()))
		}
	}()

	runs := make([]*entities.PuppetRun, 0)
	for rows.Next() {
		run := new(entities.PuppetRun)
		if err := rows.Scan(&run.ID, &run.Fqdn, &run.State, &run.ExecTime, &run.Runtime); err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}

func (m *mysqlImpl) GetRuns(ctx context.Context) ([]*entities.PuppetRun, error) {
	sqlStmt := `
	SELECT
		hash,
		fqdn,
		state,
		executed_at,
		runtime,
		environment
	FROM reports
	ORDER BY executed_at DESC;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_runs"))
	defer t.ObserveDuration()

	stmt, err := m.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	rows, err := stmt.QueryContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Error closing rows", slog.String(logging.KeyError, err.Error()))
		}
	}()

	runs := make([]*entities.PuppetRun, 0)
	for rows.Next() {
		run := new(entities.PuppetRun)
		if err := rows.Scan(&run.ID, &run.Fqdn, &run.State, &run.ExecTime, &run.Runtime, &run.Env); err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}

func (m *mysqlImpl) SaveRun(ctx context.Context, run *entities.PuppetReport) error {
	sqlStmt := `
	INSERT INTO reports(
	                    hash,
	                    fqdn,
	                    environment,
	                    state,
	                    yaml_file,
	                    executed_at,
	                    runtime,
	                    failed,
	                    changed,
	                    total,
	                    skipped
	                    )
	values(?,?,?,?,?,?,?,?,?,?,?);
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("save_run"))
	defer t.ObserveDuration()

	stmt, err := m.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			slog.Error("Error closing statement", slog.String(logging.KeyError, err.Error()))
		}
	}()

	_, err = stmt.ExecContext(ctx,
		run.ID,
		run.Fqdn,
		run.Env,
		run.State,
		run.ReportFilePath(),
		run.ExecTime.Time().Format(time.DateTime),
		run.Runtime.String(),
		run.Failed,
		run.Changed,
		run.Total,
		run.Skipped,
	)

	// If the error is that the hash already exists, then we can ignore it.
	sqlErr := new(mysql.MySQLError)
	ok := errors.As(err, &sqlErr)
	if ok && sqlErr.Number == 1062 {
		slog.Warn("Hash already exists, ignoring", slog.String(logging.KeyHash, run.ID))
		return ErrDuplicate
	} else if err != nil {
		return fmt.Errorf("error executing statement: %w", err)
	}
	return nil
}

func (m *mysqlImpl) Ping(ctx context.Context) error {
	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("ping"))
	defer t.ObserveDuration()

	if err := m.client.PingContext(ctx); err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}
	return nil
}

func (m *mysqlImpl) setup() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	sqlStmt := `
CREATE TABLE IF NOT EXISTS reports
(
    id          INTEGER PRIMARY KEY AUTO_INCREMENT,
    hash        text NOT NULL UNIQUE,
    fqdn        text,
    environment text,
    state       text,
    yaml_file   text,
    runtime     text,
    executed_at DATETIME,
    total       integer,
    skipped     integer,
    failed      integer,
    changed     integer
)
`

	stmt, err := m.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}

func NewMySQL(v *viper.Viper) (Database, error) {
	connectionString := v.GetString("db.conn_str")
	if connectionString != "" {
		slog.Debug("Found MySQL URI in environment")
	} else {
		// Missing environment variable.
		return nil, fmt.Errorf("no %s environment variable provided", EnvDbConnStr)
	}

	d, err := sqlx.Open("mysql", connectionString)
	if err != nil {
		return nil, fmt.Errorf("error opening mysql: %w", err)
	}

	newDb := NewDb(d)

	impl := &mysqlImpl{
		client: newDb,
	}

	if err := impl.setup(); err != nil {
		return nil, fmt.Errorf("error setting up database: %w", err)
	}

	return impl, nil
}
