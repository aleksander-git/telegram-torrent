package database

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

type DatabaseInterface interface {
	AddTorrent(userName string, link string) error
	GetTorrents(userName string) (TorrentList, error)
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
