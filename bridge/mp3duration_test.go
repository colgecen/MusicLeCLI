package bridge

import (
	"os"
	"path/filepath"
	"testing"
)

// syntheticMP3 builds n MPEG1 Layer-III frames (128kbps, 44.1kHz).
func syntheticMP3(n int) []byte {
	const frameLen = 417 // 144*128000/44100, no padding
	out := make([]byte, 0, n*frameLen)
	for i := 0; i < n; i++ {
		f := make([]byte, frameLen)
		f[0], f[1], f[2], f[3] = 0xff, 0xfb, 0x90, 0xc0
		out = append(out, f...)
	}
	return out
}

func TestMP3DurationSec(t *testing.T) {
	p := filepath.Join(t.TempDir(), "s.mp3")
	if err := os.WriteFile(p, syntheticMP3(100), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := MP3DurationSec(p)
	if err != nil {
		t.Fatalf("sure okunamadi: %v", err)
	}
	want := 100 * 1152.0 / 44100.0 // ≈2.61s
	if got < want-0.05 || got > want+0.05 {
		t.Errorf("sure=%v, beklenen≈%v", got, want)
	}
}

func TestMP3DurationSecID3(t *testing.T) {
	// ID3v2 başlıklı dosya da okunmalı.
	tag := []byte{'I', 'D', '3', 4, 0, 0, 0, 0, 0, 20}
	tag = append(tag, make([]byte, 20)...)
	p := filepath.Join(t.TempDir(), "s.mp3")
	if err := os.WriteFile(p, append(tag, syntheticMP3(50)...), 0644); err != nil {
		t.Fatal(err)
	}
	got, err := MP3DurationSec(p)
	if err != nil {
		t.Fatalf("sure okunamadi: %v", err)
	}
	want := 50 * 1152.0 / 44100.0
	if got < want-0.05 || got > want+0.05 {
		t.Errorf("sure=%v, beklenen≈%v", got, want)
	}
}

func TestMP3DurationSecGarbage(t *testing.T) {
	p := filepath.Join(t.TempDir(), "s.mp3")
	_ = os.WriteFile(p, []byte("bu bir mp3 degil"), 0644)
	if _, err := MP3DurationSec(p); err == nil {
		t.Error("bozuk dosyada hata bekleniyordu")
	}
}
