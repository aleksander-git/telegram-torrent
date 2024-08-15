package backend

import (
	"context"
	"database/sql"
)

type Database struct {
	db      *sql.DB
	Queries *Queries // Ваша сгенерированная структура запросов
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

func (d *Database) AddUser(ctx context.Context, arg AddUserParams) error {
	return d.Queries.AddUser(ctx, arg)
}

func (d *Database) GetUser(ctx context.Context, id int64) (User, error) {
	return d.Queries.GetUser(ctx, id)
}

func (d *Database) AddTorrent(ctx context.Context, arg AddTorrentParams) error {
	return d.Queries.AddTorrent(ctx, arg)
}

func (d *Database) GetTorrent(ctx context.Context, torrentLink string) (Torrent, error) {
	return d.Queries.GetTorrent(ctx, torrentLink)
}

func (d *Database) GetTorrents(ctx context.Context) ([]Torrent, error) {
	return d.Queries.GetUserTorrents(ctx, 0)
}

func (d *Database) GetSetting(ctx context.Context, key string) (string, error) {
	return d.Queries.GetSetting(ctx, GetSettingParams{})
}
