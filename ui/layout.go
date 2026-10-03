package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// TextWidth returns the display-cell width of s (CJK runes count double,
// ANSI codes are ignored). Use it instead of len() for layout math.
func TextWidth(s string) int {
	return lipgloss.Width(s)
}

// FitLabel truncates s with an ellipsis so its display width is at most max.
// It never wraps: the result is always a single line.
func FitLabel(s string, max int) string {
	if TextWidth(s) <= max {
		return s
	}
	if max <= 0 {
		return ""
	}
	if max == 1 {
		return "…"
	}
	var b strings.Builder
	w := 0
	for _, r := range s {
		rw := TextWidth(string(r))
		if w+rw+1 > max { // +1 reserves the ellipsis cell
			break
		}
		b.WriteRune(r)
		w += rw
	}
	return b.String() + "…"
}

// FitPad truncates s with FitLabel, then right-pads it with spaces so its
// display width is exactly w. Ideal for buttons/labels whose box must keep
// a constant size in every language.
func FitPad(s string, w int) string {
	s = FitLabel(s, w)
	if pad := w - TextWidth(s); pad > 0 {
		s += strings.Repeat(" ", pad)
	}
	return s
}

// FitPadCenter is FitPad but centers s: leftover cells split left/right
// (odd cell goes right). Display-width aware, so CJK labels center correctly
// too.
func FitPadCenter(s string, w int) string {
	s = FitLabel(s, w)
	if pad := w - TextWidth(s); pad > 0 {
		left := pad / 2
		right := pad - left
		s = strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
	}
	return s
}
