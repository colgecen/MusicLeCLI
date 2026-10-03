package bridge

import (
	"os"
	"path/filepath"
	"strings"

	"MusicLeCLI/state"
)

// ReconcileLibrary scans every playlist directory and registers orphan audio
// files (on disk but missing from song_list.txt). Downloads that finished
// without registration (crash, old version, manual copies) show up in the UI
// again after this runs. It is idempotent — call it at startup and after
// downloads. Returns the number of entries added.
func ReconcileLibrary() int {
	added := 0
	for _, p := range state.Current.Profiles {
		for _, pl := range p.Playlists {
			added += reconcileDir(state.Current.PlaylistDir(p.FolderName, pl.FolderName))
		}
	}
	return added
}

func reconcileDir(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	existing, _ := state.ReadSongs(filepath.Join(dir, "song_list.txt"))
	have := make(map[string]bool, len(existing))
	for _, s := range existing {
		have[s.Filename] = true
	}
	added := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !audioExts[strings.ToLower(filepath.Ext(e.Name()))] {
			continue
		}
		if have[e.Name()] {
			continue
		}
		full := filepath.Join(dir, e.Name())
		meta := extractMetadata(full)
		// Downloads carry ID3 tags, but be lenient: split
		// "Artist - Title.mp3" filenames when tags are absent.
		if meta.Artist == "" || meta.Artist == "Unknown" {
			if a, t := splitArtistTitle(e.Name()); a != "" && t != "" {
				meta.Artist, meta.Title = a, t
			}
		}
		registerSong(dir, full, meta)
		added++
	}
	return added
}

// splitArtistTitle splits "Artist - Title.mp3" into its parts.
func splitArtistTitle(filename string) (artist, title string) {
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	parts := strings.SplitN(base, " - ", 2)
	if len(parts) != 2 {
		return "", strings.TrimSpace(base)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}
