package loader_test

import (
	"context"
    "fmt"
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
			uri:     "magnet:?xt=urn:btih:27f3930fb49568be40ca7f572f89cf2c36f946a3&dn=rainforest+(3).jpg&tr=wss%3A%2F%2Ftracker.btorrent.xyz&tr=wss%3A%2F%2Ftracker.openwebtorrent.com&tr=udp%3A%2F%2Ftracker.leechers-paradise.org%3A6969&tr=udp%3A%2F%2Ftracker.coppersurfer.tk%3A6969&tr=udp%3A%2F%2Ftracker.opentrackr.org%3A1337&tr=udp%3A%2F%2Fexplodie.org%3A6969&tr=udp%3A%2F%2Ftracker.empire-js.us%3A1337",
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

			size, err := l.Load(ctx, test.uri, 2 * time.Second, func(_ context.Context, totalBytes, bytesCompleted int64) {
				fmt.Printf ("downloaded %d bytes of %d\n", bytesCompleted, totalBytes)
            })
			require.NoError(t, err, "failed to download file from uri %q", test.uri)

			fmt.Printf("successfully downloaded %d bytes\n", size)

			errs := client.Close()
			if len(errs) > 0 {
				t.Fatalf("errors while closing client: %+v", errs)
			}
		})
	}
}
