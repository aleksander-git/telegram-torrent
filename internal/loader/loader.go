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

func (l *Loader) Load(ctx context.Context, magnetUri string) error {
	const src = "Loader.Load"
	log := l.log.With(slog.String("src", src))
	log.Debug("loading torrent...", slog.String("uri", magnetUri))

	t, err := l.client.AddMagnet(magnetUri)
	if err != nil {
		return fmt.Errorf("l.client.AddMagnet(%q): %w", magnetUri, err)
	}

	ctx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

	for processing := true; processing ; {
		select {
		case <-ctx.Done():
			return fmt.Errorf("failed to get info: %w", ctx.Err())
		case <-t.GotInfo():
			fmt.Println("got info", t.Info().Length)
			processing = false
		}
	}

	t.DownloadAll()

	totalBytes := t.Info().Length
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("file loading failed: %w", ctx.Err())
		case <-ticker.C:
			bytesCompleted := t.BytesCompleted()

			if bytesCompleted >= totalBytes {
				log.Debug("torrent loaded",
					slog.String("uri", magnetUri),
					slog.Int64("size", totalBytes),
				)
				return nil
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
