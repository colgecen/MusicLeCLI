package bridge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dhowden/tag"
)

var supportedAudioExt = map[string]string{
	".mp3":  "MP3",
	".flac": "FLAC",
	".m4a":  "AAC",
	".mp4":  "AAC",
	".aac":  "AAC",
	".ogg":  "OGG",
	".wav":  "WAV",
	".opus": "Opus",
}

// extractMetadata reads audio file tags and returns a Result with title, artist, album, etc.
func extractMetadata(filePath string) *Result {
	base := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
	ext := strings.ToLower(filepath.Ext(filePath))
	result := &Result{
		Status:   "ok",
		Filename: filepath.Base(filePath),
		Title:    base,
		Artist:   "Unknown",
		Album:    "",
		Format:   supportedAudioExt[ext],
	}

	f, err := os.Open(filePath)
	if err != nil {
		result.Status = "error"
		result.Error = fmt.Sprintf("open file: %v", err)
		return result
	}
	defer f.Close()

	meta, err := tag.ReadFrom(f)
	if err != nil {
		// Return basic result without tags
		return result
	}

	if title := meta.Title(); title != "" {
		result.Title = title
	}
	if artist := meta.Artist(); artist != "" {
		result.Artist = artist
	}
	if album := meta.Album(); album != "" {
		result.Album = album
	}

	// Playing time for MP3 files (best effort; 0 when unreadable).
	if ext == ".mp3" {
		if secs, err := MP3DurationSec(filePath); err == nil && secs > 0 {
			result.Duration = secs
		}
	}

	// NOTE: embedded pictures are intentionally NOT extracted to _art sidecar
	// files anymore. Covers are read straight from the audio files when
	// needed (now-playing card), so no stray image folders are created.
	// Stale _art / playlist_avatar folders are removed by CleanUnusedArtwork.

	return result
}
