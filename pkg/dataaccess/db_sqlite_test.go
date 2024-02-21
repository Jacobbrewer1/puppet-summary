package dataaccess

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Jacobbrewer1/puppet-summary/pkg/entities"
	"github.com/stretchr/testify/suite"
)

type sqliteSuite struct {
	suite.Suite

	// db is the database connection.
	db *sql.DB

	// mockDB is the mock database connection.
	mockDB sqlmock.Sqlmock

	// dbObject is the database object.
	dbObject *sqliteImpl
}

func TestSqliteSuite(t *testing.T) {
	suite.Run(t, new(sqliteSuite))
}

func (s *sqliteSuite) SetupTest() {
	// Create a mock database connection.
	db, mock, err := sqlmock.New()
	s.Require().NoError(err)
	s.db = db
	s.mockDB = mock

	// Create a new database object.
	s.dbObject = &sqliteImpl{
		client: db,
	}
}

func (s *sqliteSuite) TearDownTest() {
	s.db = nil
	s.mockDB = nil
	s.dbObject = nil
}

func (s *sqliteSuite) TestClose() {
	// Expect the database to be closed.
	s.mockDB.ExpectClose()

	err := s.dbObject.Close(context.Background())
	s.Require().NoError(err)
}

func (s *sqliteSuite) TestPing() {
	// Expect the database to be pinged.
	s.mockDB.ExpectPing()

	err := s.dbObject.Ping(context.Background())
	s.Require().NoError(err)
}

func (s *sqliteSuite) TestPurge() {
	expSql := regexp.QuoteMeta(`
		DELETE FROM reports
		WHERE executed_at < ?;
	`)

	from, err := time.Parse(time.DateTime, "2023-02-21 00:00:00")
	s.Require().NoError(err)

	// Expect the database to be purged.
	s.mockDB.ExpectPrepare(expSql)
	s.mockDB.ExpectExec(expSql).
		WithArgs(from.Format(time.DateTime)).
		WillReturnResult(sqlmock.NewResult(0, 5))

	affected, err := s.dbObject.Purge(context.Background(), from)
	s.Require().NoError(err)

	s.Require().Equal(5, affected)
}

func (s *sqliteSuite) TestGetEnvironments() {
	expSql := regexp.QuoteMeta(`
		SELECT DISTINCT environment
		FROM reports;
	`)

	// Expect the environments to be retrieved.
	s.mockDB.ExpectPrepare(expSql)

	rows := sqlmock.NewRows([]string{"environment"}).
		AddRow("PRODUCTION").
		AddRow("STAGING").
		AddRow("DEVELOPMENT")

	s.mockDB.ExpectQuery(expSql).
		WillReturnRows(rows)

	s.mockDB.ExpectClose()

	environments, err := s.dbObject.GetEnvironments(context.Background())
	s.Require().NoError(err)

	s.Require().Equal([]entities.Environment{
		entities.EnvProduction,
		entities.EnvStaging,
		entities.EnvDevelopment,
	}, environments)
}

func (s *sqliteSuite) TestGetHistoryAllEnvs() {
	expSql1 := regexp.QuoteMeta(`SELECT DISTINCT DATE(executed_at) FROM reports;`)

	// Expect the history to be retrieved.
	s.mockDB.ExpectPrepare(expSql1)

	rows1 := sqlmock.NewRows([]string{"DATE(executed_at)"}).
		AddRow("2023-02-21").
		AddRow("2023-02-22").
		AddRow("2023-02-23")

	s.mockDB.ExpectQuery(expSql1).
		WillReturnRows(rows1)

	expSql2 := regexp.QuoteMeta(`SELECT DISTINCT state, COUNT('state') FROM reports WHERE executed_at BETWEEN ? AND ? GROUP BY state;`)

	from1, err := time.Parse(time.DateTime, "2023-02-21 00:00:00")
	s.Require().NoError(err)

	to1, err := time.Parse(time.DateTime, "2023-02-22 00:00:00")
	s.Require().NoError(err)

	// Expect the history to be retrieved.
	s.mockDB.ExpectPrepare(expSql2)

	rows2 := sqlmock.NewRows([]string{"state", "COUNT('state')"}).
		AddRow("CHANGED", 5).
		AddRow("FAILURE", 1).
		AddRow("UNCHANGED", 3)

	s.mockDB.ExpectQuery(expSql2).
		WithArgs(from1.Format(time.DateOnly), to1.Format(time.DateOnly)).
		WillReturnRows(rows2)

	from2, err := time.Parse(time.DateTime, "2023-02-22 00:00:00")
	s.Require().NoError(err)

	to2, err := time.Parse(time.DateTime, "2023-02-23 00:00:00")
	s.Require().NoError(err)

	// Expect the history to be retrieved.
	s.mockDB.ExpectPrepare(expSql2)

	rows3 := sqlmock.NewRows([]string{"state", "COUNT('state')"}).
		AddRow("CHANGED", 3).
		AddRow("FAILURE", 0).
		AddRow("UNCHANGED", 6)

	s.mockDB.ExpectQuery(expSql2).
		WithArgs(from2.Format(time.DateOnly), to2.Format(time.DateOnly)).
		WillReturnRows(rows3)

	from3, err := time.Parse(time.DateTime, "2023-02-23 00:00:00")
	s.Require().NoError(err)

	to3, err := time.Parse(time.DateTime, "2023-02-24 00:00:00")
	s.Require().NoError(err)

	// Expect the history to be retrieved.
	s.mockDB.ExpectPrepare(expSql2)

	rows4 := sqlmock.NewRows([]string{"state", "COUNT('state')"}).
		AddRow("CHANGED", 2).
		AddRow("FAILURE", 0).
		AddRow("UNCHANGED", 7)

	s.mockDB.ExpectQuery(expSql2).
		WithArgs(from3.Format(time.DateOnly), to3.Format(time.DateOnly)).
		WillReturnRows(rows4)

	s.mockDB.ExpectClose()

	history, err := s.dbObject.GetHistory(context.Background(), entities.EnvAll)
	s.Require().NoError(err)

	s.Require().Equal([]*entities.PuppetHistory{
		{
			Date:      "2023-02-21",
			Changed:   5,
			Failed:    0,
			Unchanged: 3,
		},
		{
			Date:      "2023-02-22",
			Changed:   3,
			Failed:    0,
			Unchanged: 6,
		},
		{
			Date:      "2023-02-23",
			Changed:   2,
			Failed:    0,
			Unchanged: 7,
		},
	}, history)
}

func (s *sqliteSuite) TestGetHistorySingleEnv() {
	expSql1 := regexp.QuoteMeta(`SELECT DISTINCT DATE(executed_at) FROM reports;`)

	// Expect the history to be retrieved.
	s.mockDB.ExpectPrepare(expSql1)

	rows1 := sqlmock.NewRows([]string{"DATE(executed_at)"}).
		AddRow("2023-02-21").
		AddRow("2023-02-22").
		AddRow("2023-02-23")

	s.mockDB.ExpectQuery(expSql1).
		WillReturnRows(rows1)

	expSql2 := regexp.QuoteMeta(`SELECT DISTINCT state, COUNT('state') FROM reports WHERE executed_at BETWEEN ? AND ? AND environment = 'PRODUCTION' GROUP BY state;`)

	from1, err := time.Parse(time.DateTime, "2023-02-21 00:00:00")
	s.Require().NoError(err)

	to1, err := time.Parse(time.DateTime, "2023-02-22 00:00:00")
	s.Require().NoError(err)

	// Expect the history to be retrieved.
	s.mockDB.ExpectPrepare(expSql2)

	rows2 := sqlmock.NewRows([]string{"state", "COUNT('state')"}).
		AddRow("CHANGED", 5).
		AddRow("FAILURE", 1).
		AddRow("UNCHANGED", 3)

	s.mockDB.ExpectQuery(expSql2).
		WithArgs(from1.Format(time.DateOnly), to1.Format(time.DateOnly)).
		WillReturnRows(rows2)

	from2, err := time.Parse(time.DateTime, "2023-02-22 00:00:00")
	s.Require().NoError(err)

	to2, err := time.Parse(time.DateTime, "2023-02-23 00:00:00")
	s.Require().NoError(err)

	// Expect the history to be retrieved.
	s.mockDB.ExpectPrepare(expSql2)

	rows3 := sqlmock.NewRows([]string{"state", "COUNT('state')"}).
		AddRow("CHANGED", 3).
		AddRow("FAILURE", 0).
		AddRow("UNCHANGED", 6)

	s.mockDB.ExpectQuery(expSql2).
		WithArgs(from2.Format(time.DateOnly), to2.Format(time.DateOnly)).
		WillReturnRows(rows3)

	from3, err := time.Parse(time.DateTime, "2023-02-23 00:00:00")
	s.Require().NoError(err)

	to3, err := time.Parse(time.DateTime, "2023-02-24 00:00:00")
	s.Require().NoError(err)

	// Expect the history to be retrieved.
	s.mockDB.ExpectPrepare(expSql2)

	rows4 := sqlmock.NewRows([]string{"state", "COUNT('state')"}).
		AddRow("CHANGED", 2).
		AddRow("FAILURE", 0).
		AddRow("UNCHANGED", 7)

	s.mockDB.ExpectQuery(expSql2).
		WithArgs(from3.Format(time.DateOnly), to3.Format(time.DateOnly)).
		WillReturnRows(rows4)

	s.mockDB.ExpectClose()

	history, err := s.dbObject.GetHistory(context.Background(), entities.EnvProduction)
	s.Require().NoError(err)

	s.Require().Equal([]*entities.PuppetHistory{
		{
			Date:      "2023-02-21",
			Changed:   5,
			Failed:    0,
			Unchanged: 3,
		},
		{
			Date:      "2023-02-22",
			Changed:   3,
			Failed:    0,
			Unchanged: 6,
		},
		{
			Date:      "2023-02-23",
			Changed:   2,
			Failed:    0,
			Unchanged: 7,
		},
	}, history)
}

func (s *sqliteSuite) TestGetReport() {
	expSql := regexp.QuoteMeta(`
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
	`)

	id := "hash"
	ctx := context.Background()

	// Expect the report to be retrieved.
	s.mockDB.ExpectPrepare(expSql)

	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "fqdn", "environment", "state", "executed_at", "runtime", "failed", "changed", "total", "yaml_file"}).
		AddRow(id, "fqdn", "PRODUCTION", "CHANGED", now, "10s", 1, 2, 3, "yaml_file")

	s.mockDB.ExpectQuery(expSql).
		WithArgs(id).
		WillReturnRows(rows)

	s.mockDB.ExpectClose()

	report, err := s.dbObject.GetReport(ctx, id)
	s.Require().NoError(err)

	s.Require().Equal(&entities.PuppetReport{
		ID:       id,
		Fqdn:     "fqdn",
		Env:      entities.EnvProduction,
		State:    entities.StateChanged,
		ExecTime: entities.Datetime(now),
		Runtime:  entities.Duration(10 * time.Second),
		Failed:   1,
		Changed:  2,
		Total:    3,
		YamlFile: "yaml_file",
	}, report)
}

func (s *sqliteSuite) TestGetReports() {
	expSql := regexp.QuoteMeta(`
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
		yaml_file
	FROM reports
	WHERE fqdn = ?
	ORDER BY executed_at DESC;
	`)

	id1 := "hash1"
	id2 := "hash2"
	ctx := context.Background()

	// Expect the report to be retrieved.
	s.mockDB.ExpectPrepare(expSql)

	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "fqdn", "environment", "state", "executed_at", "runtime", "failed", "changed", "total", "yaml_file"}).
		AddRow(id1, "fqdn", "PRODUCTION", "CHANGED", now, "10s", 1, 2, 3, "yaml_file").
		AddRow(id2, "fqdn", "DEVELOPMENT", "CHANGED", now, "11s", 1, 2, 3, "yaml_file1")

	s.mockDB.ExpectQuery(expSql).
		WithArgs("fqdn").
		WillReturnRows(rows)

	s.mockDB.ExpectClose()

	report, err := s.dbObject.GetReports(ctx, "fqdn")
	s.Require().NoError(err)

	s.Require().Equal([]*entities.PuppetReportSummary{
		{
			ID:       id1,
			Fqdn:     "fqdn",
			Env:      entities.EnvProduction,
			State:    entities.StateChanged,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
			Failed:   1,
			Changed:  2,
			Total:    3,
			YamlFile: "yaml_file",
		},
		{
			ID:       id2,
			Fqdn:     "fqdn",
			Env:      entities.EnvDevelopment,
			State:    entities.StateChanged,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(11 * time.Second),
			Failed:   1,
			Changed:  2,
			Total:    3,
			YamlFile: "yaml_file1",
		},
	}, report)
}

func (s *sqliteSuite) TestGetRunsByStateSingleState() {
	expSql := regexp.QuoteMeta(`
			SELECT
				hash,
				fqdn,
				state,
				executed_at,
				runtime
			FROM reports
			WHERE state IN (?)
			ORDER BY executed_at DESC;
		`)

	ctx := context.Background()

	// Expect the report to be retrieved.
	s.mockDB.ExpectPrepare(expSql)

	now := time.Now()

	rows := sqlmock.NewRows([]string{"hash", "fqdn", "state", "executed_at", "runtime"}).
		AddRow("hash1", "fqdn1", "CHANGED", now, "10s").
		AddRow("hash2", "fqdn2", "CHANGED", now, "11s")

	s.mockDB.ExpectQuery(expSql).
		WithArgs("CHANGED").
		WillReturnRows(rows)

	s.mockDB.ExpectClose()

	report, err := s.dbObject.GetRunsByState(ctx, entities.StateChanged)
	s.Require().NoError(err)

	s.Require().Equal([]*entities.PuppetRun{
		{
			ID:       "hash1",
			Fqdn:     "fqdn1",
			State:    entities.StateChanged,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
		},
		{
			ID:       "hash2",
			Fqdn:     "fqdn2",
			State:    entities.StateChanged,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(11 * time.Second),
		},
	}, report)
}

func (s *sqliteSuite) TestGetRunsByStateMultipleStates() {
	expSql := regexp.QuoteMeta(`
			SELECT
				hash,
				fqdn,
				state,
				executed_at,
				runtime
			FROM reports
			WHERE state IN (?)
			ORDER BY executed_at DESC;
		`)

	ctx := context.Background()

	// Expect the report to be retrieved.
	s.mockDB.ExpectPrepare(expSql)

	now := time.Now()

	rows := sqlmock.NewRows([]string{"hash", "fqdn", "state", "executed_at", "runtime"}).
		AddRow("hash1", "fqdn1", "CHANGED", now, "10s").
		AddRow("hash2", "fqdn2", "UNCHANGED", now, "11s")

	s.mockDB.ExpectQuery(expSql).
		WithArgs("CHANGED,UNCHANGED").
		WillReturnRows(rows)

	s.mockDB.ExpectClose()

	report, err := s.dbObject.GetRunsByState(ctx, entities.StateChanged, entities.StateUnchanged)
	s.Require().NoError(err)

	s.Require().Equal([]*entities.PuppetRun{
		{
			ID:       "hash1",
			Fqdn:     "fqdn1",
			State:    entities.StateChanged,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
		},
		{
			ID:       "hash2",
			Fqdn:     "fqdn2",
			State:    entities.StateUnchanged,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(11 * time.Second),
		},
	}, report)
}

func (s *sqliteSuite) TestGetRuns() {
	expSql := regexp.QuoteMeta(`
	SELECT
		hash,
		fqdn,
		state,
		executed_at,
		runtime,
		environment
	FROM reports
	ORDER BY executed_at DESC;
	`)

	ctx := context.Background()

	// Expect the report to be retrieved.
	s.mockDB.ExpectPrepare(expSql)

	now := time.Now()

	rows := sqlmock.NewRows([]string{"hash", "fqdn", "state", "executed_at", "runtime", "environment"}).
		AddRow("hash1", "fqdn1", "CHANGED", now, "10s", "PRODUCTION").
		AddRow("hash2", "fqdn2", "UNCHANGED", now, "11s", "DEVELOPMENT")

	s.mockDB.ExpectQuery(expSql).
		WillReturnRows(rows)

	s.mockDB.ExpectClose()

	report, err := s.dbObject.GetRuns(ctx)
	s.Require().NoError(err)

	s.Require().Equal([]*entities.PuppetRun{
		{
			ID:       "hash1",
			Fqdn:     "fqdn1",
			State:    entities.StateChanged,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(10 * time.Second),
			Env:      entities.EnvProduction,
		},
		{
			ID:       "hash2",
			Fqdn:     "fqdn2",
			State:    entities.StateUnchanged,
			ExecTime: entities.Datetime(now),
			Runtime:  entities.Duration(11 * time.Second),
			Env:      entities.EnvDevelopment,
		},
	}, report)
}

func (s *sqliteSuite) TestSaveRunSuccess() {
	expSql := regexp.QuoteMeta(`
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
	`)

	ctx := context.Background()

	now, err := time.Parse(time.DateTime, "2024-02-21 10:20:53")
	s.Require().NoError(err)

	// Expect the report to be saved.
	s.mockDB.ExpectPrepare(expSql)
	s.mockDB.ExpectExec(expSql).
		WithArgs("hash", "fqdn", "PRODUCTION", "CHANGED", "reports/PRODUCTION/fqdn/2024-02-21T10:20:53Z.yaml",
			now.Format(time.DateTime), "10s", 1, 2, 3, 0).
		WillReturnResult(sqlmock.NewResult(0, 1))

	s.mockDB.ExpectClose()

	err = s.dbObject.SaveRun(ctx, &entities.PuppetReport{
		ID:       "hash",
		Fqdn:     "fqdn",
		Env:      entities.EnvProduction,
		State:    entities.StateChanged,
		YamlFile: "yaml_file",
		ExecTime: entities.Datetime(now),
		Runtime:  entities.Duration(10 * time.Second),
		Failed:   1,
		Changed:  2,
		Total:    3,
	})
	s.Require().NoError(err)
}

func (s *sqliteSuite) TestSaveRunDuplicate() {
	expSql := regexp.QuoteMeta(`
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
	`)

	ctx := context.Background()

	now, err := time.Parse(time.DateTime, "2024-02-21 10:20:53")
	s.Require().NoError(err)

	// Expect the report to be saved.
	s.mockDB.ExpectPrepare(expSql)
	s.mockDB.ExpectExec(expSql).
		WithArgs("hash", "fqdn", "PRODUCTION", "CHANGED", "reports/PRODUCTION/fqdn/2024-02-21T10:20:53Z.yaml",
			now.Format(time.DateTime), "10s", 1, 2, 3, 0).
		WillReturnError(errors.New("UNIQUE constraint failed: reports.hash"))

	s.mockDB.ExpectClose()

	err = s.dbObject.SaveRun(ctx, &entities.PuppetReport{
		ID:       "hash",
		Fqdn:     "fqdn",
		Env:      entities.EnvProduction,
		State:    entities.StateChanged,
		YamlFile: "yaml_file",
		ExecTime: entities.Datetime(now),
		Runtime:  entities.Duration(10 * time.Second),
		Failed:   1,
		Changed:  2,
		Total:    3,
	})
	s.Require().Equal(ErrDuplicate, err)
}
