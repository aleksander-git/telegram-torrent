package loader_test

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/aleksander-git/telegram-torrent/internal/loader"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load(t *testing.T) {
	l, err := loader.New(slog.New(slog.NewTextHandler(os.Stdout, nil)), "./tmp", 60*time.Second)
	require.NoError(t, err, "failed to init loader")

	err = l.Load(context.Background(), "magnet:?xt=urn:btih:dd8255ecdc7ca55fb0bbf81323d87062db1f6d1c&dn=Big+Buck+Bunny&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.fastcast.nz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&ws=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2F&xs=https%3A%2F%2Fwebtorrent.io%2Ftorrents%2Fbig-buck-bunny.torrent")
	require.NoError(t, err)
}
