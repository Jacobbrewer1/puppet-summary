package dataaccess

type DbOpt string

const (
	// DbSqlite is the sqlite database.
	DbSqlite DbOpt = "SQLITE"

	// DbMySQL is the MySQL database.
	DbMySQL DbOpt = "MYSQL"

	// DbMongo is the MongoDB database.
	DbMongo DbOpt = "MONGO"
)

func (d DbOpt) String() string {
	return string(d)
}

func (d DbOpt) IsIn(opts ...DbOpt) bool {
	for _, opt := range opts {
		if opt == d {
			return true
		}
	}
	return false
}

func (d DbOpt) Valid() bool {
	return d.IsIn(
		DbSqlite,
		DbMySQL,
		DbMongo,
	)
}

var GCSEnabled = false
