package ui

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

// Uygulama artık arka planı zorlamaz; terminalin kendi teması kullanılır.
// AppStyle, tüm ekranı siyaha boyayan bir zemin özelliğine sahip olmamalı.
func TestAppStyleDoesNotForceBackground(t *testing.T) {
	InitStyles()
	empty := lipgloss.NoColor{}
	if got := AppStyle.GetBackground(); got != empty {
		t.Fatalf("AppStyle arka planı = %v, beklenen boş (terminal teması)", got)
	}
}

func TestParseHexColor(t *testing.T) {
	ok := map[string]string{
		"#ff00aa":      "#FF00AA",
		"ff00aa":       "#FF00AA",
		"  #00e5ff  ":  "#00E5FF",
		"255,0,170":    "#FF00AA",
		" 0, 229,255":  "#00E5FF",
		"rgb(255,0,7)": "#FF0007",
	}
	for in, want := range ok {
		got, good := ParseHexColor(in)
		if !good || got != want {
			t.Errorf("ParseHexColor(%q) = %q,%v; want %q,true", in, got, good, want)
		}
	}
	for _, bad := range []string{"", "#fff", "gggggg", "1,2", "1,2,3,4", "256,0,0", "-1,0,0", "red", "#12 45 78"} {
		if _, good := ParseHexColor(bad); good {
			t.Errorf("ParseHexColor(%q) gecerli sayildi", bad)
		}
	}
}

func TestApplyThemeHexAndPreset(t *testing.T) {
	InitStyles()
	ApplyTheme("#00e5ff")
	if string(ColorAccent) != "#00E5FF" {
		t.Fatalf("hex tema sonrasi Accent = %v", ColorAccent)
	}
	ApplyTheme("hacker")
	if string(ColorAccent) != ThemeColors["hacker"] {
		t.Fatalf("preset sonrasi Accent = %v", ColorAccent)
	}
	before := ColorAccent
	ApplyTheme("boyle-bir-renk-yok")
	if ColorAccent != before {
		t.Fatalf("gecersiz tema degistirmemeliydi: %v", ColorAccent)
	}
}

func TestThemePresetsHaveSamples(t *testing.T) {
	for _, want := range []string{"neon", "hacker", "kitty", "anime", "rose", "crimson", "teal", "violet"} {
		hex, ok := ThemeColors[want]
		if !ok {
			t.Errorf("preset eksik: %q", want)
			continue
		}
		if _, ok := ParseHexColor(hex); !ok {
			t.Errorf("preset %q hex gecersiz: %q", want, hex)
		}
	}
}
