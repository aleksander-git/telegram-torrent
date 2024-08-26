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

	torrentFile, err := l.Torrent(ctx, magnetUri)
	if err != nil {
		return 0, fmt.Errorf("l.Torrent(ctx, %q): %w", magnetUri, err)
	}

	size, err := l.LoadTorrent(ctx, torrentFile, loadTickInterval, onLoadTick)
	if err != nil {
		return 0, fmt.Errorf("l.LoadTorrent(ctx, %q, %d, onLoadTick): %w", torrentFile.Info().NameUtf8, loadTickInterval, err)
	}

	return size, nil
}

func (l *Loader) LoadTorrent(
	ctx context.Context,
	torrentFile *torrent.Torrent,
	loadTickInterval time.Duration,
	onLoadTick func(ctx context.Context, totalBytes, bytesCompleted int64),
) (bytesSize int64, err error) {
	const src = "Loader.LoadTorrent"
	log := l.log.With(slog.String("src", src))

	ctx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

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
					slog.Int64("size", totalBytes),
					slog.String("name", torrentFile.Info().NameUtf8),
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

func (l *Loader) Torrent(
	ctx context.Context,
	magnetUri string,
) (*torrent.Torrent, error) {
	const src = "Loader.Torrent"
	log := l.log.With(slog.String("src", src))
	log.Debug("loading torrent info...", slog.String("uri", magnetUri))

	ctx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

	torrentFile, err := l.client.AddMagnet(magnetUri)
	if err != nil {
		return nil, fmt.Errorf("l.client.AddMagnet(%q): %w", magnetUri, err)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("failed to get info: %w", ctx.Err())
	case <-torrentFile.GotInfo():
		fmt.Println("got info", torrentFile.Info().Length)
	}

	return torrentFile, nil
}
