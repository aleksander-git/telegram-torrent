package uploader_test

import (
	"context"
	stdLog "log"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/aleksander-git/telegram-torrent/internal/uploader"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// before running tests, set env variables into "./test/test.env":
	// BOT_TOKEN, APP_ID, APP_HASH, TARGET
	if err := godotenv.Load("./test/test.env"); err != nil {
		stdLog.Fatal("failed to load env file: ", err)
	}

	os.Exit(m.Run())
}

func TestUploader_Upload(t *testing.T) {
	target := os.Getenv("TARGET")

	tests := []struct {
		name     string
		filePath string
		target   string
	}{
		{
			name:     "sending_txt_file_test",
			filePath: "./test/test_files/hello-world.txt",
			target:   target,
		}, {
			name:     "sending_audio_file_test",
			filePath: "./test/test_files/merry-xmas.mp3",
			target:   target,
		}, {
			name:     "sending_video_file_test",
			filePath: "./test/test_files/nature-video.mp4",
			target:   target,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test := test
			t.Parallel()

			// use go test --short to skip uploading of heavy files
			if testing.Short() && filepath.Ext(test.filePath) != ".txt" {
				t.Skip()
			}

			appId, err := strconv.Atoi(os.Getenv("APP_ID"))
			require.NoError(t, err, "failed to parse APP_ID")
			uploader, err := uploader.New(context.Background(), slog.Default(), uploader.BotCredentials{
				BotToken: os.Getenv("BOT_TOKEN"),
				AppID:    appId,
				AppHash:  os.Getenv("APP_HASH"),
			})
			require.NoError(t, err)

			defer func() {
				if err := uploader.Close(); err != nil {
					stdLog.Fatal("uploader.Close(): ", err)
				}
			}()

			err = uploader.Upload(context.Background(), test.filePath, test.target)
			require.NoError(t, err)
		})
	}
}
