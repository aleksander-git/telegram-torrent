package loader

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/anacrolix/torrent"
)

type Loader struct {
	log *slog.Logger

	client *torrent.Client

	timeout time.Duration
}

func New(log *slog.Logger, destPath string, timeout time.Duration) (*Loader, error) {
	cfg := torrent.NewDefaultClientConfig()
	cfg.DataDir = destPath
	client, err := torrent.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("torrent.NewClient(&torrent.ClientConfig{DataDir: %q}): %w", destPath, err)
	}

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
	fmt.Println("loading torrent...", magnetUri)

	t, err := l.client.AddMagnet(magnetUri)
	if err != nil {
		return fmt.Errorf("l.client.AddMagnet(%q): %w", magnetUri, err)
	}

	<-t.GotInfo()
	t.DownloadAll()

	ctx, cancel := context.WithTimeout(ctx, l.timeout)
	defer cancel()

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
				fmt.Println("torrent loaded", totalBytes)
				return nil
			} else {
				log.Debug("torrent loading...",
					slog.Float64("percentage", float64(bytesCompleted)/float64(totalBytes)*100),
					slog.Int64("bytesCompleted", bytesCompleted),
					slog.Int64("totalBytes", totalBytes),
				)
				fmt.Printf("loading... %.2f%% (%d/%d)\n", float64(bytesCompleted)/float64(totalBytes)*100, bytesCompleted, totalBytes)
			}
		}
	}
}

func (l *Loader) Close() error {
	if errs := l.client.Close(); len(errs) > 0 {
		return fmt.Errorf("l.client.Close(): %w", errors.Join(errs...))
	}
	return nil
}
