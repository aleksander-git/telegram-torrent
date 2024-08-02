package loader

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/anacrolix/torrent"
)

type Loader struct {
	log *slog.Logger

	client TorrentClient

	timeout time.Duration
}

type TorrentClient interface {
	AddMagnet(uri string) (T *torrent.Torrent, err error)
}

func New(log *slog.Logger, client TorrentClient, timeout time.Duration) (*Loader, error) {
	return &Loader{
		log:     log,
		client:  client,
		timeout: timeout,
	}, nil
}

func (l *Loader) Load(
	ctx context.Context,
	magnetUri string,
	loadTickInterval time.Duration,
	onLoadTick func(ctx context.Context, totalBytes, bytesCompleted int64),
) (bytesSize int64, err error) {
	const src = "Loader.Load"
	log := l.log.With(slog.String("src", src))
	log.Debug("loading torrent...", slog.String("uri", magnetUri))

	torrentFile, err := l.client.AddMagnet(magnetUri)
	if err != nil {
		return 0, fmt.Errorf("l.client.AddMagnet(%q): %w", magnetUri, err)
	}

	ctx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

	for processing := true; processing ; {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("failed to get info: %w", ctx.Err())
		case <-torrentFile.GotInfo():
			fmt.Println("got info", torrentFile.Info().Length)
			processing = false
		}
	}

	torrentFile.DownloadAll()

	totalBytes := torrentFile.Info().Length
	ticker := time.NewTicker(loadTickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return 0, fmt.Errorf("file loading failed: %w", ctx.Err())
		case <-ticker.C:
			bytesCompleted := torrentFile.BytesCompleted()

			if onLoadTick != nil {
				onLoadTick(ctx, totalBytes, bytesCompleted)
			}

			if bytesCompleted >= totalBytes {
				log.Debug("torrent loaded",
					slog.String("uri", magnetUri),
					slog.Int64("size", totalBytes),
				)
				return totalBytes, nil
			} else {
				log.Debug("torrent loading...",
					slog.Float64("percentage", float64(bytesCompleted)/float64(totalBytes)*100),
					slog.Int64("bytesCompleted", bytesCompleted),
					slog.Int64("totalBytes", totalBytes),
				)
			}
		}
	}
}
