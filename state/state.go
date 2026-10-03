package state

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Language codes
type Language string

const (
	LangEnglish    Language = "en"
	LangTurkish    Language = "tr"
	LangSpanish    Language = "es"
	LangGerman     Language = "de"
	LangFrench     Language = "fr"
	LangArabic     Language = "ar"
	LangPortuguese Language = "pt"
	LangChinese    Language = "zh"
	LangJapanese   Language = "ja"
	LangItalian    Language = "it"
	LangRussian    Language = "ru"
)

// LanguageEndonym returns the endonym (native name) for each language code.
func LanguageEndonym(lang Language) string {
	switch lang {
	case LangEnglish:
		return "English"
	case LangTurkish:
		return "Türkçe"
	case LangSpanish:
		return "Español"
	case LangGerman:
		return "Deutsch"
	case LangFrench:
		return "Français"
	case LangArabic:
		return "العربية"
	case LangPortuguese:
		return "Português"
	case LangChinese:
		return "中文"
	case LangJapanese:
		return "日本語"
	case LangItalian:
		return "Italiano"
	case LangRussian:
		return "Русский"
	}
	return "English"
}

// AllLanguages returns every supported language in display order.
func AllLanguages() []Language {
	return []Language{
		LangTurkish,
		LangEnglish,
		LangSpanish,
		LangGerman,
		LangFrench,
		LangArabic,
		LangPortuguese,
		LangChinese,
		LangJapanese,
		LangItalian,
		LangRussian,
	}
}

// T returns localized text (en or tr)
func T(lang Language, en, tr string) string {
	if lang == LangTurkish {
		return tr
	}
	return en
}

// Song represents a single audio track
type Song struct {
	Filename  string
	Title     string
	Artist    string
	DateAdded string
	Duration  string
	FilePath  string
	Source    string // download source (videoId/watch URL or search query)
}

// Playlist represents a named collection of songs
type Playlist struct {
	FolderName string
	Name       string
	Bio        string
	ArtPath    string
	Songs      []Song
	IsPrivate  bool
	CreatedAt  string
}

// Profile represents a user profile
type Profile struct {
	FolderName  string
	DisplayName string
	Bio         string
	Language    Language
	Playlists   []Playlist
}

// PlayerState tracks audio playback state
type PlayerState struct {
	IsPlaying   bool
	IsPaused    bool
	CurrentSong *Song
	Position    float64 // seconds
	Duration    float64 // seconds
	Volume      float64 // 0.0-1.0
	IsShuffled  bool
	IsPrivate   bool
	StatusMsg   string
	IsError     bool
	Format      string      // e.g. "MP3", "FLAC"
	SampleRate  int         // e.g. 44100
	Bitrate     int         // e.g. 320 (kbps)
	AudioLevelL float64     // 0.0-1.0 VU left
	AudioLevelR float64     // 0.0-1.0 VU right
	Spectrum    [17]float64 // frequency spectrum
}

// AppState is the central singleton state
type AppState struct {
	RootDir         string
	Language        Language
	Profiles        []Profile
	CurrentProfile  *Profile
	CurrentPlaylist *Playlist
	Player          PlayerState
	IsFirstLaunch   bool
	ConfigDir       string
	NetworkOnline   bool
	Theme           string // accent color theme name

	// Sound settings
	SoundOutputDevice string // selected output sink name (empty = system default)
	SoundVolumeLimit  int    // max percentage of device volume the app may use (0-100)

	// Spectrum visualizer color palette name (see ui.SpectrumPalettes)
	SpectrumPalette string
}

// Current is the global app state
var Current = &AppState{
	Player:           PlayerState{Volume: 0.7},
	Language:         LangEnglish,
	Theme:            "green",
	SoundVolumeLimit: 100,
	SpectrumPalette:  "RGB",
}

// savedConfig is the on-disk persistent config format
type savedConfig struct {
	RootDir           string   `json:"root_dir"`
	Language          Language `json:"language"`
	LastUser          string   `json:"last_user"`
	Theme             string   `json:"theme"`
	SoundOutputDevice string   `json:"sound_output_device"`
	SoundVolumeLimit  int      `json:"sound_volume_limit"`
	SpectrumPalette   string   `json:"spectrum_palette"`
}

func (a *AppState) configPath() string {
	return filepath.Join(a.ConfigDir, "config.json")
}

// LoadConfig reads config.json from disk
func (a *AppState) LoadConfig() error {
	data, err := os.ReadFile(a.configPath())
	if err != nil {
		return err
	}
	var cfg savedConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return err
	}
	a.RootDir = cfg.RootDir
	a.Language = cfg.Language
	a.Theme = cfg.Theme
	if a.Theme == "" {
		a.Theme = "green"
	}
	a.SoundOutputDevice = cfg.SoundOutputDevice
	a.SoundVolumeLimit = cfg.SoundVolumeLimit
	if a.SoundVolumeLimit <= 0 || a.SoundVolumeLimit > 100 {
		a.SoundVolumeLimit = 100
	}
	a.SpectrumPalette = cfg.SpectrumPalette
	if a.SpectrumPalette == "" {
		a.SpectrumPalette = "RGB"
	}
	return nil
}

// SaveConfig writes config.json to disk
func (a *AppState) SaveConfig() error {
	if err := os.MkdirAll(a.ConfigDir, 0755); err != nil {
		return err
	}
	cfg := savedConfig{
		RootDir:           a.RootDir,
		Language:          a.Language,
		Theme:             a.Theme,
		SoundOutputDevice: a.SoundOutputDevice,
		SoundVolumeLimit:  a.SoundVolumeLimit,
		SpectrumPalette:   a.SpectrumPalette,
	}
	if a.CurrentProfile != nil {
		cfg.LastUser = a.CurrentProfile.FolderName
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(a.configPath(), data, 0644)
}

// InitializeBaseDirs creates the root music and profiles directory early
func (a *AppState) InitializeBaseDirs(rootDir string) error {
	a.RootDir = rootDir
	return os.MkdirAll(a.ProfilesDir(), 0755)
}

// ProfilesDir returns the profiles/ root directory
func (a *AppState) ProfilesDir() string {
	return filepath.Join(a.RootDir, "profiles")
}

// ScanProfiles re-scans the profiles/ directory and populates a.Profiles
func (a *AppState) ScanProfiles() error {
	dir := a.ProfilesDir()
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	a.Profiles = nil
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		p, err := loadProfile(filepath.Join(dir, e.Name()), e.Name())
		if err == nil {
			a.Profiles = append(a.Profiles, p)
		}
	}
	return nil
}

func loadProfile(dir, folderName string) (Profile, error) {
	p := Profile{
		FolderName:  folderName,
		DisplayName: readTxt(filepath.Join(dir, "name.txt"), folderName),
		Bio:         readTxt(filepath.Join(dir, "bio.txt"), ""),
		Language:    Language(readTxt(filepath.Join(dir, "lang.txt"), "en")),
	}
	// Scan playlists
	plDir := filepath.Join(dir, "playlists")
	if entries, err := os.ReadDir(plDir); err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			pl, err := LoadPlaylist(filepath.Join(plDir, e.Name()), e.Name())
			if err == nil {
				p.Playlists = append(p.Playlists, pl)
			}
		}
	}
	return p, nil
}

// LoadPlaylist loads a playlist from its directory
func LoadPlaylist(dir, folderName string) (Playlist, error) {
	pl := Playlist{
		FolderName: folderName,
		Name:       readTxt(filepath.Join(dir, "playlist_name.txt"), folderName),
		Bio:        readTxt(filepath.Join(dir, "playlist_bio.txt"), ""),
	}
	pl.CreatedAt = readTxt(filepath.Join(dir, "playlist_created.txt"), "")
	if pl.CreatedAt == "" {
		if fi, err := os.Stat(dir); err == nil {
			pl.CreatedAt = fi.ModTime().Format("2006-01-02 15:04")
		}
	}
	pl.ArtPath = findArtFile(dir)
	if pl.ArtPath == "" {
		pl.ArtPath = findArtFile(filepath.Join(dir, "playlist_art"))
	}
	pl.Songs = parseSongList(filepath.Join(dir, "song_list.txt"), dir)
	return pl, nil
}

func findArtFile(dir string) string {
	if entries, err := os.ReadDir(dir); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(e.Name()))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				return filepath.Join(dir, e.Name())
			}
		}
	}
	return ""
}

func parseSongList(listPath, plDir string) []Song {
	data, err := os.ReadFile(listPath)
	if err != nil {
		return nil
	}
	var songs []Song
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 5 {
			continue
		}
		s := Song{
			Filename:  parts[0],
			Title:     parts[1],
			Artist:    parts[2],
			DateAdded: parts[3],
			Duration:  normalizeDuration(parts[4]),
			FilePath:  filepath.Join(plDir, parts[0]),
		}
		if len(parts) == 6 {
			s.Source = parts[5]
		}
		songs = append(songs, s)
	}
	return songs
}

// normalizeDuration ensures a song duration is rendered as mm:ss. Values that
// are already in mm:ss form are passed through; bare integers (raw seconds,
// e.g. legacy song_list.txt entries) are converted so durations are always
// stored and shown as dakika:saniye.
func normalizeDuration(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "00:00"
	}
	if strings.Contains(s, ":") {
		return s
	}
	secs, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	return fmt.Sprintf("%02d:%02d", secs/60, secs%60)
}

// CreateProfileStructure writes the full directory/file scaffold for a new profile
func (a *AppState) CreateProfileStructure(folderName, displayName, bio string, lang Language) error {
	profileDir := filepath.Join(a.ProfilesDir(), folderName)
	for _, d := range []string{profileDir, filepath.Join(profileDir, "playlists")} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	if err := writeTxt(filepath.Join(profileDir, "name.txt"), displayName); err != nil {
		return err
	}
	if err := writeTxt(filepath.Join(profileDir, "bio.txt"), bio); err != nil {
		return err
	}
	if err := writeTxt(filepath.Join(profileDir, "lang.txt"), string(lang)); err != nil {
		return err
	}
	return nil
}

// CreatePlaylistStructure writes the full directory/file scaffold for a new playlist
func (a *AppState) CreatePlaylistStructure(profileFolder, plFolder, plName, plBio, artSrc string) error {
	plDir := filepath.Join(a.ProfilesDir(), profileFolder, "playlists", plFolder)
	if err := os.MkdirAll(plDir, 0755); err != nil {
		return err
	}
	if err := writeTxt(filepath.Join(plDir, "playlist_name.txt"), plName); err != nil {
		return err
	}
	if err := writeTxt(filepath.Join(plDir, "playlist_bio.txt"), plBio); err != nil {
		return err
	}
	if artSrc != "" {
		ext := filepath.Ext(artSrc)
		if ext == "" {
			ext = ".jpg"
		}
		_ = CopyFile(artSrc, filepath.Join(plDir, "art"+ext))
	}
	return nil
}

// AppendSong appends a song entry line to song_list.txt
func AppendSong(listPath, filename, title, artist, duration string) error {
	return AppendSongSource(listPath, filename, title, artist, duration, "")
}

// AppendSongSource appends a song entry line (including its download Source)
// to song_list.txt.
func AppendSongSource(listPath, filename, title, artist, duration, source string) error {
	entry := strings.Join([]string{filename, title, artist, time.Now().Format("2006-01-02"), duration, source}, "|") + "\n"
	f, err := os.OpenFile(listPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(entry)
	return err
}

// SaveProfileMeta updates name.txt and bio.txt
func (a *AppState) SaveProfileMeta(folderName, displayName, bio string) error {
	dir := filepath.Join(a.ProfilesDir(), folderName)
	if err := writeTxt(filepath.Join(dir, "name.txt"), displayName); err != nil {
		return err
	}
	return writeTxt(filepath.Join(dir, "bio.txt"), bio)
}

// SavePlaylistMeta updates playlist_name.txt and playlist_bio.txt
func (a *AppState) SavePlaylistMeta(profileFolder, plFolder, name, bio string) error {
	dir := filepath.Join(a.ProfilesDir(), profileFolder, "playlists", plFolder)
	if err := writeTxt(filepath.Join(dir, "playlist_name.txt"), name); err != nil {
		return err
	}
	return writeTxt(filepath.Join(dir, "playlist_bio.txt"), bio)
}

// DeletePlaylist removes a playlist directory entirely
func (a *AppState) DeletePlaylist(profileFolder, plFolder string) error {
	return os.RemoveAll(filepath.Join(a.ProfilesDir(), profileFolder, "playlists", plFolder))
}

// SongListPath returns the path to song_list.txt for the given playlist
func (a *AppState) SongListPath(profileFolder, plFolder string) string {
	return filepath.Join(a.ProfilesDir(), profileFolder, "playlists", plFolder, "song_list.txt")
}

// PlaylistDir returns the directory for the given playlist
func (a *AppState) PlaylistDir(profileFolder, plFolder string) string {
	return filepath.Join(a.ProfilesDir(), profileFolder, "playlists", plFolder)
}

// ReadSongs reads all song entries from song_list.txt
func ReadSongs(listPath string) ([]Song, error) {
	data, err := os.ReadFile(listPath)
	if err != nil {
		return nil, err
	}
	var songs []Song
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 6)
		if len(parts) < 5 {
			continue
		}
		s := Song{
			Filename:  parts[0],
			Title:     parts[1],
			Artist:    parts[2],
			DateAdded: parts[3],
			Duration:  normalizeDuration(parts[4]),
		}
		if len(parts) == 6 {
			s.Source = parts[5]
		}
		songs = append(songs, s)
	}
	return songs, nil
}

// WriteSongs writes all song entries to song_list.txt
func WriteSongs(listPath string, songs []Song) error {
	var buf strings.Builder
	for _, s := range songs {
		buf.WriteString(strings.Join([]string{s.Filename, s.Title, s.Artist, s.DateAdded, normalizeDuration(s.Duration), s.Source}, "|"))
		buf.WriteByte('\n')
	}
	return os.WriteFile(listPath, []byte(buf.String()), 0644)
}

// RemoveSong removes a song entry from song_list.txt by filename and deletes
// the audio file from the playlist folder. The file must resolve inside the
// list's directory — path traversal outside is refused. A file that is
// already gone is not an error; the entry is still removed.
func RemoveSong(listPath, filename string) error {
	songs, err := ReadSongs(listPath)
	if err != nil {
		return fmt.Errorf("read song list: %w", err)
	}
	found := false
	filtered := songs[:0]
	for _, s := range songs {
		if s.Filename == filename {
			found = true
		} else {
			filtered = append(filtered, s)
		}
	}
	if !found {
		return fmt.Errorf("song not found: %s", filename)
	}
	if err := WriteSongs(listPath, filtered); err != nil {
		return err
	}
	dir := filepath.Dir(listPath)
	target := filepath.Join(dir, filename)
	if rel, err := filepath.Rel(dir, target); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("refused to delete outside playlist dir: %s", filename)
	}
	if err := os.Remove(target); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	return nil
}

// UpdateSong updates title, artist, and/or duration of a song entry. Empty fields are left unchanged.
func UpdateSong(listPath, filename, title, artist, duration string) error {
	songs, err := ReadSongs(listPath)
	if err != nil {
		return fmt.Errorf("read song list: %w", err)
	}
	found := false
	for i, s := range songs {
		if s.Filename == filename {
			if title != "" {
				songs[i].Title = title
			}
			if artist != "" {
				songs[i].Artist = artist
			}
			if duration != "" {
				songs[i].Duration = normalizeDuration(duration)
			}
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("song not found: %s", filename)
	}
	return WriteSongs(listPath, songs)
}

// ---- helpers ----

func readTxt(path, def string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return def
	}
	return strings.TrimSpace(string(data))
}

func writeTxt(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// CopyFile copies src file to dst (creates parent dirs)
func CopyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()
	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return err
}
