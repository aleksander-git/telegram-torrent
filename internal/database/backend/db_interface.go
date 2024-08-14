package backend

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type DBInterface interface {
	AddUser(ctx context.Context, arg AddUserParams) error
	GetUser(ctx context.Context, id int64) (User, error)
	AddTorrent(ctx context.Context, arg AddTorrentParams) error
	GetTorrent(ctx context.Context, torrentLink string) (Torrent, error)
}

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
