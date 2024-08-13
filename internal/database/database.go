package database

import (
	"database/sql"
	"errors"
)

// заглушка для базы данных
// в дальнейшем всё перепишется под базу данных
type Database struct {
	cache map[string]TorrentList
}

func New() *Database {
	return &Database{
		cache: make(map[string]TorrentList),
	}
}

func (db *Database) AddTorrent(userName string, link string) error {
	db.cache[userName] = append(db.cache[userName], Torrent{
		Link:   link,
		Status: InQueue,
	})
	return nil
}

func (db *Database) GetTorrents(userName string) (TorrentList, error) {
	return db.cache[userName], nil
}

func (db *Database) GetSetting(setting string, userID sql.NullInt64) (int64, error) {
	if setting == "channel" {
		return -1002184825487, nil
	}

	return 0, errors.New("unknown setting")
}
