package backend

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"
)

type Database struct {
	Queries *Queries
	db      *sql.DB
}

func NewDatabase(connectionString string) (*Database, error) {
	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	// Проверка соединения с базой данных
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Database{
		Queries: New(db),
		db:      db,
	}, nil
}

// Метод для закрытия соединения с базой данных
func (d *Database) Close() error {
	return d.db.Close()
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

func (d *Database) GetTorrents(ctx context.Context, userID int64) ([]Torrent, error) {
	return d.Queries.GetUserTorrents(ctx, userID)
}

func (d *Database) GetSetting(ctx context.Context, userID int64, key string) (string, error) {
	params := GetSettingParams{
		UserID: sql.NullInt64{Int64: userID, Valid: true},
		Name:   key,
	}
	return d.Queries.GetSetting(ctx, params)
}
