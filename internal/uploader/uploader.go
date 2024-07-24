package uploader

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"text/template"

	"github.com/aleksander-git/telegram-torrent/utils/files"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	tduploader "github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
)

type Uploader struct {
	log *slog.Logger

	messageTemplate string
}

func New(log *slog.Logger) *Uploader {
	return &Uploader{
		log:             log,
		messageTemplate: "",
	}
}

// WithMessage adds a default message to an every sended file
//
// You can use template variables and functions or html tags like <i> or <b> to pretty message
//
// Available template variables: {{.FileName}}, {{.Extension}}, {{.IsVideo}}, {{.IsAudio}}
func (u *Uploader) WithMessage(messageTemplate string) *Uploader {
	u.messageTemplate = messageTemplate

	return u
}

// Upload uploads file located on the filePath to the Telegram server
// and sends it to targetDomain (channel name or username)
func (u *Uploader) Upload(ctx context.Context, filePath string, targetDomain string) (err error) {
	const src = "Uploader.Upload"
	log := u.log.With(
		slog.String("src", src),
	)

	performUpload := func(ctx context.Context, client *telegram.Client) error {
		api := tg.NewClient(client)
		tgUploader := tduploader.NewUploader(api)
		sender := message.NewSender(api).WithUploader(tgUploader)

		log.Debug("uploading file",
			slog.String("path", filePath),
		)

		upload, err := tgUploader.FromPath(ctx, filePath)
		if err != nil {
			return fmt.Errorf("upload %q: %w", filePath, err)
		}

		var msg string
		if u.messageTemplate != "" {
			msg, err = parseMessageTemplate(u.messageTemplate, filePath)
			if err != nil {
				return fmt.Errorf("message template parse error: %w", err)
			}
		}

		document := message.UploadedDocument(upload,
			html.String(nil, msg),
		)

		var extension = filepath.Ext(filePath)

		document.
			MIME(files.GetMimeTypeByExtension(extension)).
			Filename(filepath.Base(filePath))

		if files.IsAudio(extension) {
			document.Audio()
		} else if files.IsVideo(extension) {
			document.Video()
		}

		target := sender.Resolve(targetDomain)

		log.Debug("sending file",
			slog.String("path", filePath),
		)

		if _, err := target.Media(ctx, document); err != nil {
			return fmt.Errorf("send: %w", err)
		}

		return nil
	}
	return telegram.BotFromEnvironment(ctx, telegram.Options{
		NoUpdates: true,
	}, nil, performUpload)
}

type BotConfigs struct {
	// Token, which you can get when creating a new bot in https://t.me/BotFather
	BotToken string

	// AppID and AppHash are need to work with github.com/gotd/td library.
	// With them lib can use a MTProto API for uploading big files
	// You can get them after registering a new app on https://my.telegram.org
	AppID   string
	AppHash string
}

// LoadConfigs creates environment variables for working with library
//
// BOT_TOKEN: your bot token
//
// APP_ID: app api id
//
// APP_HASH: app api hash
//
// You can get AppID and AppHash on https://my.telegram.org
func LoadConfigs(cfg *BotConfigs) {
	os.Setenv("BOT_TOKEN", cfg.BotToken)
	os.Setenv("APP_ID", cfg.AppID)
	os.Setenv("APP_HASH", cfg.AppHash)
}

func parseMessageTemplate(messageTemplate, filePath string) (str string, err error) {
	templ, err := template.New("message").Parse(messageTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer

	err = templ.Execute(&buf, struct {
		FileName  string
		Extension string
		IsVideo   bool
		IsAudio   bool
	}{
		FileName:  filepath.Base(filePath),
		Extension: filepath.Ext(filePath),
		IsVideo:   files.IsVideo(filepath.Ext(filePath)),
		IsAudio:   files.IsAudio(filepath.Ext(filePath)),
	})

	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
