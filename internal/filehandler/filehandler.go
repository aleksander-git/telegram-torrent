package filehandler

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path"
	"time"

	"github.com/aleksander-git/telegram-torrent/internal/database/backend"
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
	) (messageID int64, err error)
}

type TorrentProvider interface {
	GetFirstUnstartedTorrent(
		ctx context.Context,
	) (backend.Torrent, error)
}

type TorrentStateUpdater interface {
	UpdateTorrentMessageID(
		ctx context.Context,
		arg backend.UpdateTorrentMessageIDParams,
	) error

	UpdateTorrentName(
		ctx context.Context,
		arg backend.UpdateTorrentNameParams,
	) error

	UpdateTorrentSize(
		ctx context.Context,
		arg backend.UpdateTorrentSizeParams,
	) error

	UpdateTorrentStatus(
		ctx context.Context,
		arg backend.UpdateTorrentStatusParams,
	) error
}

type FileHandler struct {
	log *slog.Logger

	loader   Loader
	uploader Uploader

	torrentProvider TorrentProvider
	torrentUpdater  TorrentStateUpdater

	loadTickInterval time.Duration
	onLoadTick       func(ctx context.Context, magnetUri string, totalBytes, bytesCompleted int64)

	loadDir      string
	targetDomain string

	done chan struct{}
}

func New(
	log *slog.Logger,
	loader Loader,
	uploader Uploader,
	torrentProvider TorrentProvider,
	torrentUpdater TorrentStateUpdater,

	loadTickInterval time.Duration,
	onLoadTick func(ctx context.Context, magnetUri string, totalBytes, bytesCompleted int64),
	loadDir string,
	targetDomain string,
) *FileHandler {
	return &FileHandler{
		log:              log,
		loader:           loader,
		uploader:         uploader,
		torrentProvider:  torrentProvider,
		torrentUpdater:   torrentUpdater,
		loadTickInterval: loadTickInterval,
		onLoadTick:       onLoadTick,
		loadDir:          loadDir,
		targetDomain:     targetDomain,
		done:             make(chan struct{}),
	}
}

func (h *FileHandler) Run(ctx context.Context, scanInterval time.Duration) {
	const src = "FileHandler.Run"
	log := h.log.With(slog.String("src", src))

	ticker := time.NewTicker(scanInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Error("context error received", slog.String("error", ctx.Err().Error()))
				return
			case <-h.done:
				log.Debug("shutdown filehandler")
				return
			case <-ticker.C:
				go func() {
					err := h.handle(ctx)
					if err != nil {
						log.Error("handle error received", slog.String("error", err.Error()))
					}
				}()
			}
		}
	}()
}

func (h *FileHandler) Close() {
	close(h.done)
}

func (h *FileHandler) handle(ctx context.Context) (err error) {
	const src = "FileHandler.Handle"
	log := h.log.With(slog.String("src", src))

	var magnetUri string

	defer func() {
		if err != nil && magnetUri != "" {
			statusErr := h.torrentError(ctx, magnetUri, err)
			if statusErr != nil {
				err = fmt.Errorf("h.torrentError(ctx, %q, %w): %w", magnetUri, err, statusErr)
			}
		}
	}()

	unstartedTorrent, err := h.torrentProvider.GetFirstUnstartedTorrent(ctx)
	if err != nil {
		return fmt.Errorf("h.torrentProvider.GetFirstUnstartedTorrent(ctx): %w", err)
	}

	magnetUri = unstartedTorrent.TorrentLink

	log.Debug("handle torrent", slog.String("link", magnetUri))

	torrentFile, err := h.loader.Torrent(ctx, magnetUri)
	if err != nil {
		return fmt.Errorf("h.loader.Torrent(ctx, %q): %w", magnetUri, err)
	}

	err = h.preLoadUpdateTorrentState(
		ctx,
		magnetUri,
		torrentFile.Info().BestName(),
		torrentFile.Info().Length,
	)

	if err != nil {
		return fmt.Errorf("failed to update torrent %q status: %w", magnetUri, err)
	}

	_, err = h.loader.LoadTorrent(
		ctx,
		torrentFile,
		h.loadTickInterval,
		func(ctx context.Context, totalBytes, bytesCompleted int64) {
			h.onLoadTick(ctx, magnetUri, totalBytes, bytesCompleted)
		},
	)

	if err != nil {
		return fmt.Errorf("failed to load torrent: %w", err)
	}

	filePath := path.Join(h.loadDir, torrentFile.Info().BestName())
	messageID, err := h.uploader.Upload(ctx, filePath, h.targetDomain)
	if err != nil {
		return fmt.Errorf("h.uploader.Upload(ctx, %q, %q): %w", filePath, h.targetDomain, err)
	}

	err = h.postLoadUpdateTorrentState(ctx, magnetUri, messageID)
	if err != nil {
		return fmt.Errorf("failed to torrent %q status: %w", magnetUri, err)
	}

	return nil
}

func (h *FileHandler) preLoadUpdateTorrentState(
	ctx context.Context,
	magnerUri string,
	name string,
	size int64,
) error {
	err := h.torrentUpdater.UpdateTorrentName(ctx, backend.UpdateTorrentNameParams{
		TorrentLink: magnerUri,
		Name:        sql.NullString{String: name, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("h.torrentUpdater.UpdateTorrentName(ctx, %q, %q): %w", magnerUri, name, err)
	}

	err = h.torrentUpdater.UpdateTorrentSize(ctx, backend.UpdateTorrentSizeParams{
		TorrentLink: magnerUri,
		Size:        sql.NullInt64{Int64: size, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("h.torrentUpdater.UpdateTorrentSize(ctx, %q, %d): %w", magnerUri, size, err)
	}

	err = h.torrentUpdater.UpdateTorrentStatus(ctx, backend.UpdateTorrentStatusParams{
		TorrentLink: magnerUri,
		TimeStarted: sql.NullTime{Time: time.Now(), Valid: true},
	})

	if err != nil {
		return fmt.Errorf("h.torrentUpdater.UpdateTorrentStatus(ctx, %q, %q): %w", magnerUri, time.Now().String(), err)
	}

	return nil
}

func (h *FileHandler) postLoadUpdateTorrentState(
	ctx context.Context,
	magnerUri string,
	messageID int64,
) error {
	err := h.torrentUpdater.UpdateTorrentMessageID(ctx, backend.UpdateTorrentMessageIDParams{
		TorrentLink: magnerUri,
		MessageID:   sql.NullInt64{Int64: messageID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("h.torrentUpdater.UpdateTorrentMessageID(ctx, %q, %d): %w", magnerUri, messageID, err)
	}

	err = h.torrentUpdater.UpdateTorrentStatus(ctx, backend.UpdateTorrentStatusParams{
		TorrentLink:  magnerUri,
		TimeFinished: sql.NullTime{Time: time.Now(), Valid: true},
	})

	if err != nil {
		return fmt.Errorf("h.torrentUpdater.UpdateTorrentStatus(ctx, %q, %q): %w", magnerUri, time.Now().String(), err)
	}

	return nil
}

func (h *FileHandler) torrentError(ctx context.Context, magnetUri string, torrentErr error) error {
	if torrentErr != nil {
		err := h.torrentUpdater.UpdateTorrentStatus(ctx, backend.UpdateTorrentStatusParams{
			TorrentLink: magnetUri,
			Error:       sql.NullString{String: torrentErr.Error(), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("h.torrentUpdater.UpdateTorrentStatus(ctx, %q, %w): %w", magnetUri, torrentErr, err)
		}
	}

	return nil
}
