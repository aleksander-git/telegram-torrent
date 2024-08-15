package backend

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type Database struct {
	*Queries
	db *sql.DB
}

func NewDatabase(connectionString string) (*Database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	return &Database{
		Queries: New(db),
		db:      db,
	}, nil
}
