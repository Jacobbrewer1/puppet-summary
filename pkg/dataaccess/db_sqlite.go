package dataaccess

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/Jacobbrewer1/puppet-summary/pkg/logging"
	_ "github.com/mattn/go-sqlite3"
	"github.com/prometheus/client_golang/prometheus"
)

func connectSQLite() {
	dbLite, err := sql.Open("sqlite3", "file:puppet-summary.db?_journal_mode=WAL")
	if err != nil {
		slog.Error("Error connecting to SQLite", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	l := slog.Default().With(slog.String(logging.KeyDal, "sqlite"))

	impl := &sqliteImpl{
		l:      l,
		client: dbLite,
	}

	if err := impl.setup(); err != nil {
		slog.Error("Error setting up SQLite", slog.String(logging.KeyError, err.Error()))
		os.Exit(1)
	}

	DB = impl

	slog.Debug("Connected to SQLite")
}

type sqliteImpl struct {
	// l is the logger.
	l *slog.Logger

	// client is the database.
	client *sql.DB
}

func (s *sqliteImpl) Close(_ context.Context) error {
	return s.client.Close()
}

func (s *sqliteImpl) Purge(ctx context.Context, from entities.Datetime) (int, error) {
	sqlStmt := `
	DELETE FROM reports
	WHERE executed_at < ?;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("purge"))
	defer t.ObserveDuration()

	stmt, err := s.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return 0, fmt.Errorf("error preparing statement: %w", err)
	}

	res, err := stmt.ExecContext(ctx, from.Time().Format(time.DateTime))
	if err != nil {
		return 0, fmt.Errorf("error executing statement: %w", err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("error getting rows affected: %w", err)
	}

	return int(rows), nil
}

func (s *sqliteImpl) GetEnvironments(ctx context.Context) ([]entities.Environment, error) {
	sqlStmt := `
	SELECT DISTINCT environment 
	FROM reports;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_environments"))
	defer t.ObserveDuration()

	stmt, err := s.client.PrepareContext(ctx, sqlStmt)
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

	envs := make([]entities.Environment, 0)
	for rows.Next() {
		var env entities.Environment
		if err := rows.Scan(&env); err != nil {
			return nil, fmt.Errorf("error scanning rows: %w", err)
		}
		envs = append(envs, env)
	}

	return envs, nil
}

func (s *sqliteImpl) GetHistory(ctx context.Context, environment entities.Environment) ([]*entities.PuppetHistory, error) {
	res := make([]*entities.PuppetHistory, 0)

	limit := 30

	query := "SELECT DISTINCT DATE(executed_at) FROM reports"
	if environment != entities.EnvAll {
		query = fmt.Sprintf("%s WHERE environment = '%s'", query, environment)
	}

	stmt, err := s.client.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	rows, err := stmt.QueryContext(ctx)
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
		dt, err := time.Parse(time.DateOnly, d)
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
		if environment != entities.EnvAll {
			locQuery += " AND environment = '" + environment.String() + "' "
		}
		locQuery += " GROUP BY state;"
		stmt, err = s.client.PrepareContext(ctx, locQuery)
		if err != nil {
			return nil, fmt.Errorf("error preparing statement: %w", err)
		}

		rows, err = stmt.QueryContext(ctx, startTime.Format(time.DateOnly), endTime.Format(time.DateOnly))
		if err != nil {
			return nil, fmt.Errorf("error executing statement: %w", err)
		}

		for rows.Next() {
			var state entities.State
			var count int

			err = rows.Scan(&state, &count)
			if err != nil {
				return nil, errors.New("failed to scan SQL")
			}
			if state.IsIn(entities.StateChanged) {
				x.Changed += count
			}
			if state.IsIn(entities.StateUnchanged) {
				x.Unchanged += count
			}
			if state.IsIn(entities.StateFailed) {
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

func (s *sqliteImpl) GetReport(ctx context.Context, id string) (*entities.PuppetReport, error) {
	sqlStmt := `
	SELECT
		hash,
		fqdn,
		environment,
		state,
		executed_at,
		runtime,
		failed,
		changed,
		total,
		skipped
	FROM reports
	WHERE hash = ?;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_report"))
	defer t.ObserveDuration()

	stmt, err := s.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	row := stmt.QueryRowContext(ctx, id)

	report := new(entities.PuppetReport)
	if err := row.Scan(&report.ID, &report.Fqdn, &report.Env, &report.State, &report.ExecTime, &report.Runtime,
		&report.Failed, &report.Changed, &report.Total, &report.Skipped); err != nil {
		return nil, fmt.Errorf("error scanning row: %w", err)
	}

	return report, nil
}

func (s *sqliteImpl) GetReports(ctx context.Context, fqdn string) ([]*entities.PuppetReportSummary, error) {
	sqlStmt := `
	SELECT
		hash,
		fqdn,
		environment,
		state,
		executed_at,
		runtime,
		failed,
		changed,
		total,
		skipped
	FROM reports
	WHERE fqdn = ?
	ORDER BY executed_at DESC;
`

	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("get_reports"))
	defer t.ObserveDuration()

	stmt, err := s.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	rows, err := stmt.QueryContext(ctx, fqdn)
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}

	defer func() {
		if err := rows.Close(); err != nil {
			slog.Error("Error closing rows", slog.String(logging.KeyError, err.Error()))
		}
	}()

	reports := make([]*entities.PuppetReportSummary, 0)
	for rows.Next() {
		report := new(entities.PuppetReportSummary)
		if err := rows.Scan(&report.ID, &report.Fqdn, &report.Env, &report.State, &report.ExecTime, &report.Runtime,
			&report.Failed, &report.Changed, &report.Total, &report.Skipped); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		reports = append(reports, report)
	}

	return reports, nil
}

func (s *sqliteImpl) Ping(ctx context.Context) error {
	// Start the prometheus metrics.
	t := prometheus.NewTimer(DatabaseLatency.WithLabelValues("ping"))
	defer t.ObserveDuration()

	return s.client.PingContext(ctx)
}

func (s *sqliteImpl) GetRunsByState(ctx context.Context, state entities.State) ([]*entities.PuppetRun, error) {
	sqlStmt := `
	SELECT
		hash,
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

	stmt, err := s.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	rows, err := stmt.QueryContext(ctx, state)
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
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}

func (s *sqliteImpl) GetRuns(ctx context.Context) ([]*entities.PuppetRun, error) {
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

	stmt, err := s.client.PrepareContext(ctx, sqlStmt)
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
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		runs = append(runs, run)
	}

	return runs, nil
}

func (s *sqliteImpl) SaveRun(ctx context.Context, run *entities.PuppetReport) error {
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

	stmt, err := s.client.PrepareContext(ctx, sqlStmt)
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
		"", // TODO: When GCS is implemented, this will need to be updated to use the yaml file.
		run.ExecTime.Time().Format(time.DateTime),
		run.Runtime.String(),
		run.Failed,
		run.Changed,
		run.Total,
		run.Skipped,
	)
	if err != nil {
		return fmt.Errorf("error saving run: %w", err)
	}
	return nil
}

func (s *sqliteImpl) setup() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	sqlStmt := `
        CREATE TABLE IF NOT EXISTS reports (
          id          INTEGER PRIMARY KEY AUTOINCREMENT,
		  hash 	      text,
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

	stmt, err := s.client.PrepareContext(ctx, sqlStmt)
	if err != nil {
		return fmt.Errorf("error preparing statement: %w", err)
	}

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}
	return nil
}
