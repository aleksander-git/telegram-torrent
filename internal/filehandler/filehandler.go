package filehandler

import (
	"context"
	"log/slog"
	"time"

	"github.com/anacrolix/torrent"
)

type Loader interface {
	Torrent(
		ctx context.Context,
		magnetUri string,
	) (*torrent.Torrent, error)

	LoadTorrent(
		ctx context.Context,
		torrentFile *torrent.Torrent,
		loadTickInterval time.Duration,
		onLoadTick func(ctx context.Context, totalBytes, bytesCompleted int64),
	) (bytesSize int64, err error)
}

type Uploader interface {
	Upload(
		ctx context.Context,
		filePath string,
		targetDomain string,
	) (err error)
}

type FileHandler struct {
	log *slog.Logger

	loader   Loader
	uploader Uploader
}
