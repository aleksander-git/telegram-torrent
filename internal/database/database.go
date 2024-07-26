package database

import (
	"fmt"

	"github.com/aleksander-git/telegram-torrent/internal/torrent"
)

// заглушка для базы данных
// в дальнейшем всё перепишется под базу данных
var (
	cache = make(map[string]torrent.TorrentList)
)

func AddTorrent(userName string, link string) error {
	t, err := torrent.New(link)
	if err != nil {
		return fmt.Errorf("cannot add torrent to database: %w", err)
	}

	cache[userName] = append(cache[userName], t)
	return nil
}

func GetTorrents(userName string) (torrent.TorrentList, error) {
	return cache[userName], nil
}
