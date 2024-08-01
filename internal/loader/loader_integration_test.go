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
			uri:     "your_magnet",
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

			err = l.Load(ctx, test.uri)
			require.NoError(t, err, "failed to download file from uri %q", test.uri)

			errs := client.Close()
			if len(errs) > 0 {
				t.Fatalf("errors while closing client: %+v", errs)
			}
		})
	}
}
