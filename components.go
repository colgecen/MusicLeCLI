package main

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"

	"MusicLeCLI/state"
	"MusicLeCLI/ui"
)

type InputField struct {
	textinput.Model
	label string
}

func NewInputField(label, placeholder string) InputField {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = label
	ti.PromptStyle = ui.AccentStyle
	ti.TextStyle = ui.WhiteStyle
	ti.PlaceholderStyle = ui.DimStyle
	ti.Cursor.Style = lipgloss.NewStyle().
		Background(lipgloss.Color("#1DB954")).
		Foreground(lipgloss.Color("#000000"))
	ti.Width = 60
	ti.CharLimit = 200
	return InputField{Model: ti, label: label}
}

func (i InputField) FocusedView() string {
	return i.Model.View()
}

func (i InputField) BlurredView() string {
	v := i.Model.View()
	return lipgloss.NewStyle().Foreground(ui.ColorSecondary).Render(v)
}

// enToKey maps legacy English strings (langT call sites) to Tr keys so the
// whole app automatically supports all 11 languages without touching every
// call site. Leading/trailing spaces are part of the lookup to preserve
// button padding.
var enToKey = map[string]string{
	"Title":                    "home.title",
	"Artist":                   "home.artist",
	"Duration":                 "home.duration",
	"Import failed: ":          "home.import_failed",
	"Imported: ":               "home.imported",
	"Updated: ":                "home.updated",
	"Update failed: ":          "home.update_failed",
	"Deleted: ":                "home.deleted",
	"Delete failed: ":          "home.delete_failed",
	"Playlist renamed: ":       "home.playlist_renamed",
	"CONSOLE":                  "home.console",
	"SPECTRUM":                 "home.spectrum",
	"SONGS":                    "home.songs",
	"Dur.":                     "home.dur_short",
	"Operations":               "home.operations",
	"PLAYLIST":                 "home.playlist",
	"Playlist Download":        "dl.playlist_download",
	"Saved!":                   "pl.saved",
	"  v Saved!":               "profile.saved",
	"Profile":                  "profile.title",
	"  Save Profile  ":         "profile.save",
	" Profile Settings":        "profile.settings",
	"Welcome":                  "setup.welcome",
	"Hos Geldiniz":             "setup.welcome",
	"Art selected":             "pl.art_selected",
	"Image must be under 1MB":  "pl.img_size",
	"Image reset":              "pl.img_reset",
	"No profile":               "pl.no_profile",
	"Name is required":         "pl.name_required",
	"Name already exists":      "pl.name_exists",
	"Playlist created!":        "pl.created",
	"Select Art Image":         "pl.select_art",
	"Resim Sec":                "pl.select_art",
	"Image Files":              "pl.img_files",
	"Resim Dosyalari":          "pl.img_files",
	"Deleted":                  "pl.deleted",
	"No playlists":             "pl.no_playlists",
	" Playlists":               "pl.playlists_title",
	" Playlist":                "pl.playlists_title",
	" New Playlist":            "pl.new_prefix",
	"(creating new)":           "pl.creating_new",
	"(click to select image)":  "pl.click_select",
	"  Save  ":                 "pl.save",
	"  Kaydet  ":               "pl.save",
	"  Create  ":               "pl.create",
	"  Oluştur  ":              "pl.create",
	"  Playlist Sil  ":         "pl.delete_btn",
	"  Playlist Ekle  ":        "pl.add_btn",
	"  Resmi Sıfırla  ":        "pl.reset_btn",
	" Playlist Settings":       "pl.settings_title",
	" Playlist Ayarlari":       "pl.settings_title",
	" Profil Ayarlari":         "profile.settings",
	"> Play All":               "home.play_all",
	"# Shuffle":                "home.shuffle",
	"Created: ":                "home.created",
	"Isim":                     "home.title",
	"Sre.":                     "home.dur_short",
	"Islemler":                 "home.operations",
}

func langT(en, tr string) string {
	if key, ok := enToKey[en]; ok {
		v := Tr(key)
		if v != key {
			return v
		}
	}
	// Fallback for strings not yet in the dictionary (or missing translation):
	// keep the old en/tr behaviour so nothing ever renders a raw key.
	return state.T(state.Current.Language, en, tr)
}

func renderLogo() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		ui.LogoStyle.Render("Music"),
		ui.LogoAccentStyle.Render("Le"),
	)
}

type Option struct {
	Text  string
	Value string
}
