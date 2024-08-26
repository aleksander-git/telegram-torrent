package filehandler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"time"

	"github.com/aleksander-git/telegram-torrent/internal/database/backend"
	"github.com/anacrolix/torrent"
)

var (
	ErrMaxSize = errors.New("torrent size is too big")
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

type Repository interface {
	GetFirstUnstartedTorrent(
		ctx context.Context,
	) (backend.Torrent, error)

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

	repository Repository

	loadTickInterval time.Duration
	onLoadTick       func(ctx context.Context, magnetUri string, totalBytes, bytesCompleted int64)

	loadDir      string
	targetDomain string

	maxTorrentSize int64
}

func New(
	log *slog.Logger,
	loader Loader,
	uploader Uploader,
	repository Repository,

	loadTickInterval time.Duration,
	onLoadTick func(ctx context.Context, magnetUri string, totalBytes, bytesCompleted int64),
	loadDir string,
	targetDomain string,
	maxTorrentSize int64,
) *FileHandler {
	return &FileHandler{
		log:              log,
		loader:           loader,
		uploader:         uploader,
		repository:       repository,
		loadTickInterval: loadTickInterval,
		onLoadTick:       onLoadTick,
		loadDir:          loadDir,
		targetDomain:     targetDomain,
		maxTorrentSize:   maxTorrentSize,
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

func (h *FileHandler) handle(ctx context.Context) error {
	magnetUri, torrentFile, err := h.loadTorrent(ctx)
	if err != nil {
		if magnetUri != "" {
			statusErr := h.torrentError(ctx, magnetUri, err)
			if statusErr != nil {
				err = fmt.Errorf("h.torrentError(ctx, %q, %w): %w", magnetUri, err, statusErr)
			}
		}
		return fmt.Errorf("failed to load torrent: %w", err)
	}

	err = h.uploadTorrent(ctx, magnetUri, torrentFile.Info().BestName())
	if err != nil {
		statusErr := h.torrentError(ctx, magnetUri, err)
		if statusErr != nil {
			err = fmt.Errorf("h.torrentError(ctx, %q, %w): %w", magnetUri, err, statusErr)
		}
		return fmt.Errorf("failed to upload torrent: %w", err)
	}

	return nil
}

func (h *FileHandler) loadTorrent(ctx context.Context) (string, *torrent.Torrent, error) {
	unstartedTorrent, err := h.repository.GetFirstUnstartedTorrent(ctx)
	if err != nil {
		return "", nil, fmt.Errorf("h.torrentProvider.GetFirstUnstartedTorrent(ctx): %w", err)
	}

	magnetUri := unstartedTorrent.TorrentLink

	torrentFile, err := h.loader.Torrent(ctx, magnetUri)
	if err != nil {
		return magnetUri, nil, fmt.Errorf("h.loader.Torrent(ctx, %q): %w", magnetUri, err)
	}

	if torrentFile.Info().Length >= h.maxTorrentSize {
		return magnetUri, nil, ErrMaxSize
	}

	err = h.preLoadUpdateTorrentState(
		ctx,
		magnetUri,
		torrentFile.Info().BestName(),
		torrentFile.Info().Length,
	)

	if err != nil {
		return magnetUri, nil, fmt.Errorf("failed to update torrent %q status: %w", magnetUri, err)
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
		return magnetUri, nil, fmt.Errorf("failed to load torrent: %w", err)
	}

	return magnetUri, torrentFile, nil
}

func (h *FileHandler) uploadTorrent(ctx context.Context, magnetUri string, name string) error {
	filePath := path.Join(h.loadDir, name)
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
	err := h.repository.UpdateTorrentName(ctx, backend.UpdateTorrentNameParams{
		TorrentLink: magnerUri,
		Name:        sql.NullString{String: name, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("h.torrentUpdater.UpdateTorrentName(ctx, %q, %q): %w", magnerUri, name, err)
	}

	err = h.repository.UpdateTorrentSize(ctx, backend.UpdateTorrentSizeParams{
		TorrentLink: magnerUri,
		Size:        sql.NullInt64{Int64: size, Valid: true},
	})

	if err != nil {
		return fmt.Errorf("h.torrentUpdater.UpdateTorrentSize(ctx, %q, %d): %w", magnerUri, size, err)
	}

	err = h.repository.UpdateTorrentStatus(ctx, backend.UpdateTorrentStatusParams{
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
	err := h.repository.UpdateTorrentMessageID(ctx, backend.UpdateTorrentMessageIDParams{
		TorrentLink: magnerUri,
		MessageID:   sql.NullInt64{Int64: messageID, Valid: true},
	})
	if err != nil {
		return fmt.Errorf("h.torrentUpdater.UpdateTorrentMessageID(ctx, %q, %d): %w", magnerUri, messageID, err)
	}

	err = h.repository.UpdateTorrentStatus(ctx, backend.UpdateTorrentStatusParams{
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
		err := h.repository.UpdateTorrentStatus(ctx, backend.UpdateTorrentStatusParams{
			TimeFinished: sql.NullTime{Time: time.Now(), Valid: true},
			TorrentLink:  magnetUri,
			Error:        sql.NullString{String: torrentErr.Error(), Valid: true},
		})
		if err != nil {
			return fmt.Errorf("h.torrentUpdater.UpdateTorrentStatus(ctx, %q, %w): %w", magnetUri, torrentErr, err)
		}
	}

	return nil
}
