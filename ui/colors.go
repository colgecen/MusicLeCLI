package ui

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// ThemeColors maps theme names to accent hex colors
var ThemeColors = map[string]string{
	"anime":   "#FF3D81",
	"blue":    "#4488FF",
	"crimson": "#DC143C",
	"green":   "#1DB954",
	"hacker":  "#39FF14",
	"kitty":   "#FF9ECF",
	"lime":    "#A6FF00",
	"neon":    "#00E5FF",
	"orange":  "#FFA500",
	"pink":    "#FF69B4",
	"purple":  "#BB86FC",
	"red":     "#FF4444",
	"rose":    "#FF5C8A",
	"teal":    "#2DD4BF",
	"violet":  "#8B5CF6",
	"yellow":  "#FFD700",
}

var (
	ColorBackground = lipgloss.Color("#000000")
	ColorSurface    = lipgloss.Color("#181818")
	ColorBorder     = lipgloss.Color("#282828")
	ColorAccent     = lipgloss.Color("#1DB954")
	ColorPrimary    = lipgloss.Color("#FFFFFF")
	ColorSecondary  = lipgloss.Color("#B3B3B3")
	ColorError      = lipgloss.Color("#FF4444")
	ColorOrange     = lipgloss.Color("#FFA500")
	ColorSuccess    = lipgloss.Color("#1DB954")
	ColorRowHover   = lipgloss.Color("#1ED760")
	ColorBlack      = lipgloss.Color("#000000")
	ColorSelection  = lipgloss.Color("#1E3223")

	// Brand colors for the browser connector — fixed, independent of theme.
	ColorSpotify      = lipgloss.Color("#1DB954") // Spotify green
	ColorSpotifyFocus = lipgloss.Color("#39FF14") // phosphor green (focused)
	ColorYouTube      = lipgloss.Color("#FF0000") // YouTube red
	ColorYouTubeFocus = lipgloss.Color("#FF3131") // phosphor red (focused)
	// Lighter variants used for the home sidebar connect cards so the brand
	// colors read clearly against the dark surface. Kept vivid (not washed out)
	// so the cards stay legible and punchy.
	ColorSpotifyLight = lipgloss.Color("#1ED760")
	ColorYouTubeLight = lipgloss.Color("#FF2D2D")
)

// Styles are mutable — call ApplyTheme() to rebuild
var (
	AppStyle            lipgloss.Style
	AccentStyle         lipgloss.Style
	WhiteStyle          lipgloss.Style
	DimStyle            lipgloss.Style
	ErrorStyle          lipgloss.Style
	OrangeStyle         lipgloss.Style
	BoldStyle           lipgloss.Style
	HeaderStyle         lipgloss.Style
	LogoStyle           lipgloss.Style
	LogoAccentStyle     lipgloss.Style
	BorderStyle         lipgloss.Style
	AccentBorderStyle   lipgloss.Style
	InputStyle          lipgloss.Style
	FocusedInputStyle   lipgloss.Style
	SelectedRowStyle    lipgloss.Style
	NavActiveStyle      lipgloss.Style
	NavInactiveStyle    lipgloss.Style
	ButtonStyle         lipgloss.Style
	AccentButtonStyle   lipgloss.Style
	ErrorButtonStyle    lipgloss.Style
	SectionTitleStyle   lipgloss.Style
	SeparatorStyle      string
	SurfaceStyle        lipgloss.Style
	FaintStyle          lipgloss.Style
	GreenDotStyle       string
	DimDotStyle         string
	SongNumStyle        lipgloss.Style
	SelectedBgStyle     lipgloss.Style
	FocusedButtonStyle  lipgloss.Style
	FocusedOutlineStyle lipgloss.Style
)

// InitStyles builds all styles with the current ColorAccent
func InitStyles() {
	AppStyle = lipgloss.NewStyle()

	AccentStyle = lipgloss.NewStyle().Foreground(ColorAccent)
	WhiteStyle = lipgloss.NewStyle().Foreground(ColorPrimary)
	DimStyle = lipgloss.NewStyle().Foreground(ColorSecondary)
	ErrorStyle = lipgloss.NewStyle().Foreground(ColorError)
	OrangeStyle = lipgloss.NewStyle().Foreground(ColorOrange)
	BoldStyle = lipgloss.NewStyle().Bold(true)

	HeaderStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Bold(true)

	LogoStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorPrimary)

	LogoAccentStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorAccent)

	BorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder)

	AccentBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent)

	InputStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Padding(0, 1)

	FocusedInputStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Border(lipgloss.NormalBorder()).
		BorderForeground(ColorAccent).
		Padding(0, 1)

	SelectedRowStyle = lipgloss.NewStyle().
		Background(ColorSelection).
		Foreground(ColorPrimary).
		Bold(true)

	NavActiveStyle = lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorBlack).
		Bold(true).
		Padding(0, 2)

	NavInactiveStyle = lipgloss.NewStyle().
		Foreground(ColorPrimary).
		Padding(0, 2)

	ButtonStyle = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 2)

	AccentButtonStyle = lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorBlack).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorAccent).
		Padding(0, 2)

	ErrorButtonStyle = lipgloss.NewStyle().
		Background(ColorError).
		Foreground(ColorPrimary).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorError).
		Padding(0, 2)

	FocusedButtonStyle = lipgloss.NewStyle().
		Background(ColorAccent).
		Foreground(ColorBlack).
		Bold(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 2)

	FocusedOutlineStyle = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 2)

	SectionTitleStyle = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Padding(0, 1)

	SeparatorStyle = lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render(strings.Repeat("-", 40))

	SurfaceStyle = lipgloss.NewStyle().
		Padding(0, 1)

	FaintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666"))

	GreenDotStyle = lipgloss.NewStyle().
		Foreground(ColorAccent).
		Render("o")

	DimDotStyle = lipgloss.NewStyle().
		Foreground(ColorBorder).
		Render("o")

	SongNumStyle = lipgloss.NewStyle().
		Foreground(ColorSecondary).
		Width(3).
		Align(lipgloss.Right)

	SelectedBgStyle = lipgloss.NewStyle().
		Background(ColorSelection)
}

// hexToRGB parses "#RRGGBB" into channels. ok is false for anything else.
func hexToRGB(h string) (r, g, b int, ok bool) {
	if len(h) != 7 || h[0] != '#' {
		return 0, 0, 0, false
	}
	var v uint
	if _, err := fmt.Sscanf(h[1:], "%06x", &v); err != nil {
		return 0, 0, 0, false
	}
	return int(v >> 16), int((v >> 8) & 0xff), int(v & 0xff), true
}

// mixHex blends two "#RRGGBB" colors; t=0 gives a, t=1 gives b.
func mixHex(a, b string, t float64) string {
	ar, ag, ab, ok := hexToRGB(a)
	if !ok {
		return b
	}
	br, bg, bb, ok := hexToRGB(b)
	if !ok {
		return a
	}
	m := func(x, y int) int {
		v := float64(x) + (float64(y)-float64(x))*t
		if v < 0 {
			v = 0
		}
		if v > 255 {
			v = 255
		}
		return int(v + 0.5)
	}
	return fmt.Sprintf("#%02X%02X%02X", m(ar, br), m(ag, bg), m(ab, bb))
}

// ParseHexColor normalizes "#RRGGBB", "RRGGBB", "r,g,b" and "rgb(r,g,b)"
// into "#RRGGBB". It reports false for anything else.
func ParseHexColor(s string) (string, bool) {
	t := strings.TrimSpace(s)
	if strings.HasPrefix(strings.ToLower(t), "rgb(") && strings.HasSuffix(t, ")") {
		t = t[4 : len(t)-1]
	}
	if strings.HasPrefix(t, "#") {
		t = t[1:]
	}
	if strings.Contains(t, ",") {
		parts := strings.Split(t, ",")
		if len(parts) != 3 {
			return "", false
		}
		var rgb [3]int
		for i, p := range parts {
			var v int
			if _, err := fmt.Sscanf(strings.TrimSpace(p), "%d", &v); err != nil || v < 0 || v > 255 {
				return "", false
			}
			rgb[i] = v
		}
		return fmt.Sprintf("#%02X%02X%02X", rgb[0], rgb[1], rgb[2]), true
	}
	if len(t) != 6 {
		return "", false
	}
	for _, c := range t {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return "", false
		}
	}
	return "#" + strings.ToUpper(t), true
}

// ApplyTheme updates the full palette from the accent color and rebuilds
// all styles. It accepts a preset name from ThemeColors or a custom
// "#RRGGBB" hex (stored as-is in config). Borders, surfaces, selection,
// muted text and success all derive from the accent, so the whole UI —
// not just the borders — follows the theme. Error red and brand colors
// stay fixed. Unknown names leave the current theme untouched.
func ApplyTheme(name string) {
	hex, ok := ThemeColors[name]
	if !ok {
		if h, ok := ParseHexColor(name); ok {
			hex = h
		} else {
			return
		}
	}
	ColorAccent = lipgloss.Color(hex)
	// Borders stay subtle like before (near-invisible dark gray with only
	// a hint of the theme); saturated borders looked heavy, especially
	// on unfocused panels.
	ColorBorder = lipgloss.Color(mixHex("#282828", hex, 0.12))
	ColorSurface = lipgloss.Color(mixHex("#0C0C0C", hex, 0.08))
	ColorSelection = lipgloss.Color(mixHex("#000000", hex, 0.25))
	ColorSecondary = lipgloss.Color(mixHex("#B3B3B3", hex, 0.15))
	ColorSuccess = lipgloss.Color(mixHex("#22C55E", hex, 0.45))
	InitStyles()
}

func VolumeColor(vol float64) lipgloss.Color {
	switch {
	case vol <= 0.33:
		return ColorAccent
	case vol <= 0.66:
		return ColorOrange
	default:
		return ColorError
	}
}

func VolumeBar(filled float64, width int) string {
	if width <= 0 {
		width = 10
	}
	n := int(filled * float64(width))
	if n < 0 {
		n = 0
	}
	if n > width {
		n = width
	}
	bar := ""
	for i := range width {
		if i < n {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

func ProgressBar(pos, dur float64, width int) string {
	if width <= 0 {
		width = 20
	}
	ratio := 0.0
	if dur > 0 {
		ratio = pos / dur
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	n := int(ratio * float64(width))
	bar := ""
	for i := 0; i < width; i++ {
		if i < n {
			bar += "-"
		} else if i == n {
			bar += "o"
		} else {
			bar += "-"
		}
	}
	return bar
}

func FormatDuration(secs float64) string {
	s := int(secs)
	return fmt.Sprintf("%02d:%02d", s/60, s%60)
}

var spectrumColors = []lipgloss.Color{
	lipgloss.Color("#FF0000"),
	lipgloss.Color("#FF3300"),
	lipgloss.Color("#FF6600"),
	lipgloss.Color("#FF9900"),
	lipgloss.Color("#FFCC00"),
	lipgloss.Color("#FFFF00"),
	lipgloss.Color("#AAFF00"),
	lipgloss.Color("#55FF00"),
	lipgloss.Color("#00FF44"),
	lipgloss.Color("#00FFAA"),
	lipgloss.Color("#00CCFF"),
	lipgloss.Color("#0099FF"),
	lipgloss.Color("#4466FF"),
	lipgloss.Color("#8833FF"),
	lipgloss.Color("#BB00FF"),
	lipgloss.Color("#FF00CC"),
	lipgloss.Color("#FF3399"),
}

// SpectrumPalettes maps a palette name to its 17 band color stops (low -> high
// frequency). Users can pick one from Settings -> Extras.
var SpectrumPalettes = map[string][]string{
	"RGB": {
		"#FF0000", "#FF3300", "#FF6600", "#FF9900", "#FFCC00", "#FFFF00",
		"#AAFF00", "#55FF00", "#00FF44", "#00FFAA", "#00CCFF", "#0099FF",
		"#4466FF", "#8833FF", "#BB00FF", "#FF00CC", "#FF3399",
	},
	"Spotify": {
		"#053B1E", "#0A5A2C", "#0E7338", "#129145", "#17B054", "#1DB954",
		"#2FD167", "#45E87C", "#39FF14", "#6BFF47", "#8CFF6B", "#A8FF8C",
		"#C4FFB0", "#DFFFCF", "#EFFFE4", "#6BFF47", "#39FF14",
	},
	"YouTube": {
		"#3B0000", "#5A0505", "#730A0A", "#941010", "#B51818", "#E01B1B",
		"#FF0000", "#FF2A2A", "#FF3131", "#FF5A5A", "#FF7A7A", "#FF9A9A",
		"#FFB0B0", "#FFC9C9", "#FFE0E0", "#FF3131", "#FF5A5A",
	},
	"Aurora": {
		"#001B3B", "#003366", "#00509E", "#0077C7", "#00A3E0", "#00C2D6",
		"#1FE0C2", "#4CE0A8", "#7CFFB0", "#4CC9FF", "#6A8CFF", "#9A6BFF",
		"#C24CFF", "#E04CE0", "#FF7CE0", "#4CC9FF", "#9A6BFF",
	},
	"Fire": {
		"#1A0000", "#3D0A00", "#5A1400", "#7A2200", "#9E3300", "#C24A00",
		"#E86A00", "#FF8C00", "#FFB300", "#FFD400", "#FFF000", "#FFFF66",
		"#FFFFB3", "#FFFFFF", "#FFE0E0", "#FFB300", "#FFD400",
	},
	"Mono": {
		"#1A1A1A", "#2B2B2B", "#3D3D3D", "#4F4F4F", "#616161", "#737373",
		"#858585", "#979797", "#A9A9A9", "#BBBBBB", "#CDCDCD", "#DFDFDF",
		"#EFEFEF", "#F7F7F7", "#FFFFFF", "#BBBBBB", "#FFFFFF",
	},
	"Ocean": {
		"#001F2E", "#00384F", "#00566E", "#00768C", "#0096A8", "#00B6C4",
		"#00D6D0", "#1FE6C8", "#4CF0D0", "#6CDFE0", "#4CA8E0", "#6A8CE0",
		"#7C7CFF", "#A06BFF", "#C99AFF", "#00D6D0", "#4CF0D0",
	},
}

// SetSpectrumPalette swaps the active spectrum colors to the named palette,
// rebuilding the style cache. Unknown names fall back to RGB.
func SetSpectrumPalette(name string) {
	src, ok := SpectrumPalettes[name]
	if !ok {
		src = SpectrumPalettes["RGB"]
	}
	spectrumColors = make([]lipgloss.Color, len(src))
	for i, h := range src {
		spectrumColors[i] = lipgloss.Color(h)
	}
	spectrumStyleCache = make([]lipgloss.Style, len(spectrumColors))
	for i, c := range spectrumColors {
		spectrumStyleCache[i] = lipgloss.NewStyle().Foreground(c)
	}
}

// SpectrumPaletteNames returns the available palette names, sorted.
func SpectrumPaletteNames() []string {
	names := make([]string, 0, len(SpectrumPalettes))
	for n := range SpectrumPalettes {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// PalettePreview renders a horizontal gradient sample of the named palette.
func PalettePreview(name string, n int) string {
	src := SpectrumPalettes[name]
	if len(src) == 0 {
		src = SpectrumPalettes["RGB"]
	}
	if n < 1 {
		n = 1
	}
	var sb strings.Builder
	for i := 0; i < n; i++ {
		idx := int(float64(i) / float64(n) * float64(len(src)))
		if idx >= len(src) {
			idx = len(src) - 1
		}
		sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(src[idx])).Render("█"))
	}
	return sb.String()
}

var waveShades = []string{" ", "░", "▒", "▓", "█"}

var spectrumStyleCache []lipgloss.Style

func init() {
	spectrumStyleCache = make([]lipgloss.Style, 17)
	for i, c := range spectrumColors {
		spectrumStyleCache[i] = lipgloss.NewStyle().Foreground(c)
	}
}

// SpectrumColor returns a Style colored for the given band index.
func SpectrumColor(band int, _ float64) lipgloss.Style {
	if band < 0 {
		band = 0
	}
	if band >= len(spectrumStyleCache) {
		band = len(spectrumStyleCache) - 1
	}
	return spectrumStyleCache[band]
}

var (
	smoothLevel float64
	smoothPeak  float64
	peakTime    time.Time
)

// VolumeBars renders N bars of █ based on audio level. Silent = nothing.
func VolumeBars(level float64, n int) string {
	if n < 1 {
		n = 1
	}
	if n > 64 {
		n = 64
	}

	amp := level
	if math.IsNaN(amp) || amp < 0 {
		amp = 0
	}
	if amp > 1 {
		amp = 1
	}

	// Smooth for snappy but not jittery response
	smoothLevel = smoothLevel*0.2 + amp*0.8

	// Peak
	if smoothLevel >= smoothPeak {
		smoothPeak = smoothLevel
	} else {
		dt := time.Since(peakTime).Seconds()
		if dt > 0.5 {
			smoothPeak *= math.Exp(-dt * 1.5)
		}
	}
	peakTime = time.Now()

	// Number of filled bars
	filled := int(smoothLevel * float64(n))
	if filled > n {
		filled = n
	}
	if filled < 0 {
		filled = 0
	}

	// Peak bar position
	peakBar := int(smoothPeak * float64(n))
	if peakBar > n {
		peakBar = n
	}
	if peakBar < 0 {
		peakBar = 0
	}

	var out strings.Builder
	for i := 0; i < n; i++ {
		if i < filled {
			if i == peakBar && peakBar >= filled && smoothPeak > smoothLevel+0.05 {
				out.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render("█"))
			} else {
				out.WriteString(lipgloss.NewStyle().Foreground(ColorAccent).Render("█"))
			}
		} else if i == peakBar && smoothPeak > 0.05 {
			out.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Render("█"))
		} else {
			out.WriteString(" ")
		}
	}
	return out.String()
}
