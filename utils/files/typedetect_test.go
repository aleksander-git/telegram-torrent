package files

import (
	"mime"
	"testing"
)

func TestGetMimeTypeByExtension(t *testing.T) {
	var extensions = []string{"", "unknown", ".mp4", ".mkv", ".avi", ".mov", ".wmv", ".webm", ".mp3", ".wav",
		".aac", ".flac", ".ogg", ".m4a", ".wma"}

	for _, ext := range extensions {
		t.Run("test extension: "+ext, func(t *testing.T) {
			expectedType := mime.TypeByExtension(ext)
			if ext == "unknown" {
				expectedType = "application/octet-stream"
			}

			res := GetMimeTypeByExtension(ext)

			if res != expectedType {
				t.Errorf("failed on extension %s:\nexpected type: %s\nbut got: %s\n", ext, expectedType, res)
			}
		})
	}
}

func TestIsVideo(t *testing.T) {
	var videoExtensions = []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm"}
	var audioExtensions = []string{".mp3", ".wav", ".aac", ".flac", ".ogg", ".m4a", ".wma"}

	for _, ext := range videoExtensions {
		t.Run("test extension: "+ext, func(t *testing.T) {
			res := IsVideo(ext)

			if res == false {
				t.Errorf("failed on extension %s: expected true", ext)
			}
		})
	}

	for _, ext := range audioExtensions {
		t.Run("test extension: "+ext, func(t *testing.T) {
			res := IsVideo(ext)

			if res == true {
				t.Errorf("failed on extension %s: expected false", ext)
			}
		})
	}
}

func TestIsAudio(t *testing.T) {
	var videoExtensions = []string{".mp4", ".mkv", ".avi", ".mov", ".wmv", ".flv", ".webm"}
	var audioExtensions = []string{".mp3", ".wav", ".aac", ".flac", ".ogg", ".m4a", ".wma"}

	for _, ext := range audioExtensions {
		t.Run("test extension: "+ext, func(t *testing.T) {
			res := IsAudio(ext)

			if res == false {
				t.Errorf("failed on extension %s: expected true", ext)
			}
		})
	}

	for _, ext := range videoExtensions {
		t.Run("test extension: "+ext, func(t *testing.T) {
			res := IsAudio(ext)

			if res == true {
				t.Errorf("failed on extension %s: expected false", ext)
			}
		})
	}
}
