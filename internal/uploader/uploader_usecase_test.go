package uploader

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	LoadConfigs(&BotConfigs{
		BotToken: "your_bot_token",
		AppID:    "your_app_api_id",
		AppHash:  "your_app_api_hash",
	})

	os.Exit(m.Run())
}

func TestUploader_Upload(t *testing.T) {
	target := "landowner7"

	tests := []struct {
		name     string
		filePath string
		target   string
	}{
		{
			name:     "sending_txt_file_test",
			filePath: "C:/Users/AV/go/src/telegram-torrent/internal/uploader/test/test_files/hello-world.txt",
			target:   target,
		}, {
			name:     "sending_audio_file_test",
			filePath: "C:/Users/AV/go/src/telegram-torrent/internal/uploader/test/test_files/merry-xmas.mp3",
			target:   target,
		}, {
			name:     "sending_video_file_test",
			filePath: "C:/Users/AV/go/src/telegram-torrent/internal/uploader/test/test_files/nature-video.mp4",
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

			uploader := New(slog.Default()).
				WithMessage(
					"{{if .IsVideo}}Your video: {{else if .IsAudio}}Your audio: {{else}} Your document with extension {{.Extension}}: {{end}} <i>{{.FileName}}</i>",
				)

			err := uploader.Upload(context.Background(), test.filePath, test.target)
			if err != nil {
				t.Error(err)
			}
		})
	}
}
