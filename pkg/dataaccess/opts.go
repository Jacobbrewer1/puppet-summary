package dataaccess

import (
	"flag"
)

var dbFlag = flag.String("db", "sqlite", "The database to use. Valid options are: sqlite, mysql, mongo")

type dbOpt string

const (
	// dbSqlite is the sqlite database.
	dbSqlite dbOpt = "SQLITE"

	// dbMySQL is the MySQL database.
	dbMySQL dbOpt = "MYSQL"

	// dbMongo is the MongoDB database.
	dbMongo dbOpt = "MONGO"
)

func (d dbOpt) String() string {
	return string(d)
}

func (d dbOpt) IsIn(opts ...dbOpt) bool {
	for _, opt := range opts {
		if opt == d {
			return true
		}
	}
	return false
}

func (d dbOpt) Valid() bool {
	return d.IsIn(
		dbSqlite,
		dbMySQL,
		dbMongo,
	)
}

var gcsFlag = flag.Bool("gcs", false, "Whether to use Google Cloud Storage")

var GCSEnabled = false
