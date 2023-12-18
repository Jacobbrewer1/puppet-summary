package connection

import (
	"database/sql"
	"errors"
	"fmt"
)

// MySQL is a connection to a mysql database.
type MySQL struct {
	// Db is the database connection.
	Db *sql.DB

	// connectionString is the connection string.
	ConnectionString string
}

// Connect connects to the database.
func (m *MySQL) Connect() error {
	if m.Db != nil {
		return errors.New("database already connected")
	}
	if m.ConnectionString == "" {
		return errors.New("no connection string provided")
	}
	db, err := sql.Open("mysql", m.ConnectionString)
	if err != nil {
		return fmt.Errorf("error opening connection to mysql: %s", err)
	}
	if err := db.Ping(); err != nil {
		if err := db.Close(); err != nil {
			return fmt.Errorf("error closing mysql: %s", err)
		}
		return fmt.Errorf("error trying to ping mysql: %s", err)
	}
	m.Db = db
	return nil
}
