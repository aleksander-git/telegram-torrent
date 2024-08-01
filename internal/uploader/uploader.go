package uploader

import (
	"bytes"
	"context"
	"fmt"
	"github.com/gotd/td/telegram/message/peer"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/tg"
)

type FileUploader interface {
	FromPath(ctx context.Context, path string) (tg.InputFileClass, error)
}

type Resolver interface {
	Resolve(from string, decorators ...peer.PromiseDecorator) *message.RequestBuilder
}

type Uploader struct {
	log *slog.Logger

	uploader FileUploader
	resolver Resolver

	messageTemplate string
}

func New(
	log *slog.Logger,
	uploader FileUploader,
	resolver Resolver,
) *Uploader {
	return &Uploader{
		log:             log,
		uploader:        uploader,
		resolver:        resolver,
		messageTemplate: "",
	}
}

// WithMessage adds a default message to an every file sent
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

	log.Debug("uploading file", slog.String("path", filePath))

	upload, err := u.uploader.FromPath(ctx, filePath)
	if err != nil {
		return fmt.Errorf("u.uploader.FromPath(ctx, %q): %w", filePath, err)
	}

	var msg string
	if u.messageTemplate != "" {
		msg, err = parseMessageTemplate(u.messageTemplate, filePath)
		if err != nil {
			return fmt.Errorf("parseMessageTemplate(%q, %q): %w", u.messageTemplate, filePath, err)
		}
	}

	var extension = filepath.Ext(filePath)
	document := message.UploadedDocument(upload, html.String(nil, msg)).
		MIME(mime.TypeByExtension(extension)).
		Filename(filepath.Base(filePath))

	switch {
	case isAudio(extension):
		document.Audio()
	case isVideo(extension):
		document.Video()
	}

	target := u.resolver.Resolve(targetDomain)

	if _, err := target.Media(ctx, document); err != nil {
		return fmt.Errorf("failed to send file %q to target %q: %w", filePath, targetDomain, err)
	}

	return nil
}

func parseMessageTemplate(messageTemplate, filePath string) (str string, err error) {
	templ, err := template.New("message").Parse(messageTemplate)
	if err != nil {
		return "", fmt.Errorf("template.New(\"message\").Parse(%q): %w", messageTemplate, err)
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
		IsVideo:   isVideo(filepath.Ext(filePath)),
		IsAudio:   isAudio(filepath.Ext(filePath)),
	})

	if err != nil {
		return "", fmt.Errorf("template.Execute(): %w", err)
	}

	return buf.String(), nil
}

func isAudio(ext string) bool {
	return commonMimeType(ext) == "audio"
}

func isVideo(ext string) bool {
	return commonMimeType(ext) == "video"
}

func commonMimeType(ext string) string {
	return strings.Split(mime.TypeByExtension(ext), "/")[0]
}
