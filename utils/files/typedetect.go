package files

import (
	"mime"
	"slices"
)

var videoExtensions = []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm"}
var audioExtensions = []string{".mp3", ".wav", ".aac", ".flac", ".ogg", ".m4a", ".wma"}

// GetMimeTypeByExtension detect file MIME-type by extension
func GetMimeTypeByExtension(extension string) string {
	if extension == "" {
		return ""
	}

	if extension[0] != '.' {
		extension = "." + extension
	}
	mimeType := mime.TypeByExtension(extension)

	if mimeType == "" {
		return "application/octet-stream"
	}

	return mimeType
}

func IsVideo(extension string) bool {
	return slices.Contains(videoExtensions, extension)
}

func IsAudio(extension string) bool {
	return slices.Contains(audioExtensions, extension)
}
