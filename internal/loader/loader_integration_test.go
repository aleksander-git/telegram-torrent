package loader_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aleksander-git/telegram-torrent/internal/loader"
	"github.com/anacrolix/torrent"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load(t *testing.T) {
	var tests = []struct {
		name     string
		dir      string
		uri      string
		fileName string
		timeout  time.Duration
	}{
		{
			name:    "download_file",
			dir:     "./tmp",
			uri:     "magnet:?xt=urn:btih:5a936459a36e5ba77a8cf3c61c13ba3b770f62c4&dn=rainforest-minimalism-5.jpg&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337",
			timeout: 20 * time.Second,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx := context.Background()

			cfg := torrent.NewDefaultClientConfig()
			cfg.DataDir = test.dir
			client, err := torrent.NewClient(cfg)
			require.NoError(t, err, "failed to init torrent client")

			l, err := loader.New(slog.Default(), client, test.timeout)
			require.NoError(t, err, "failed to init loader")

			torrentFile, err := l.Torrent(ctx, test.uri)
			require.NoError(t, err, "failed to get file from uri %q", test.uri)

			size, err := l.LoadTorrent(ctx, torrentFile, 100*time.Millisecond, func(_ context.Context, totalBytes, bytesCompleted int64) {
				t.Logf("downloaded %d bytes of %d\n", bytesCompleted, totalBytes)
			})
			require.NoError(t, err, "failed to download file from uri %q", test.uri)

			t.Logf("successfully downloaded file %s with %d bytes\n", torrentFile.Info().BestName(), size)

			errs := client.Close()
			if len(errs) > 0 {
				t.Fatalf("errors while closing client: %+v", errs)
			}
		})
	}
}
