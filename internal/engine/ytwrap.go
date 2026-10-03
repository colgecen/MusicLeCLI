package engine

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// downloadWithYTDLP drives the bundled yt-dlp + ffmpeg to turn each URL into a
// 320k MP3 with an embedded cover (APIC). All work is done by the embedded
// binaries; this package only orchestrates them and streams progress.
func downloadWithYTDLP(ext *Extracted, urls []string, outDir string, progress Progress) error {
	if progress == nil {
		progress = func(int, string) {}
	}
	if len(urls) == 0 {
		return fmt.Errorf("at least one URL is required")
	}
	for i, u := range urls {
		progress(0, fmt.Sprintf("Downloading %d/%d...", i+1, len(urls)))
		if err := downloadOne(ext, u, outDir, progress); err != nil {
			return err
		}
	}
	return nil
}

type ytdlpMeta struct {
	Title  string `json:"title"`
	Artist string `json:"artist"`
}

// downloadOne downloads a single URL: metadata + bestaudio + thumbnail via
// yt-dlp, then muxes to MP3 with the cover embedded via ffmpeg.
func downloadOne(ext *Extracted, url, outDir string, progress Progress) error {
	tmp, err := os.MkdirTemp("", "musicle_yt_")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer os.RemoveAll(tmp)

	// 1) metadata (title) for a clean output filename.
	meta := ytdlpMeta{}
	if raw, e := runYTDLPJSON(ext.YTDLP, url); e == nil && raw != "" {
		_ = json.Unmarshal([]byte(raw), &meta)
	}
	name := sanitizeFilename(meta.Title)
	if name == "" {
		name = "track"
	}

	audioTpl := filepath.Join(tmp, "audio.%(ext)s")
	thumbBase := filepath.Join(tmp, "audio")

	// 2) download bestaudio + thumbnail. YouTube occasionally serves a
	// transient 403 on the media, so we retry a few times; if the audio file
	// shows up despite a non-zero exit (e.g. only the thumbnail step failed)
	// we accept it. On each retry we rotate the YouTube player client, since a
	// throttled client signature is a common cause of repeated 403s.
	baseArgs := []string{
		"-f", "bestaudio",
		"-o", audioTpl,
		"--write-thumbnail",
		"--no-playlist",
		"--no-warnings",
		"--retries", "3",
		"--fragment-retries", "3",
	}
	// yt-dlp auto-selects a working client by default; these are fallbacks used
	// only when a retry is needed.
	clientFallbacks := []string{"android_vr", "web", "ios", "tv", "web_synthetic", "android", "mweb"}
	var dlErr error
	for attempt := 1; attempt <= 3; attempt++ {
		dlArgs := append([]string{}, baseArgs...)
		if attempt > 1 {
			client := clientFallbacks[(attempt-2)%len(clientFallbacks)]
			dlArgs = append(dlArgs, "--extractor-args", "youtube:player_client="+client)
		}
		dlArgs = append(dlArgs, "--", url)
		dlErr = runYTDLP(ext.YTDLP, dlArgs, progress)
		if dlErr == nil {
			break
		}
		if _, err := findAudioFile(tmp, "audio"); err == nil {
			dlErr = nil
			break
		}
		if attempt < 3 {
			progress(0, fmt.Sprintf("Yeniden deneniyor (%d/3)...", attempt+1))
			time.Sleep(2 * time.Second)
		}
	}
	if dlErr != nil {
		return fmt.Errorf("yt-dlp indirme: %w", dlErr)
	}

	audioFile, err := findAudioFile(tmp, "audio")
	if err != nil {
		return err
	}
	thumbFile := findThumbnail(thumbBase)
	if thumbFile != "" {
		thumbFile = ensureCoverFormat(ext.FFMPEG, thumbFile)
	}

	// 3) ffmpeg: audio -> mp3 320k, attach cover as APIC.
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return fmt.Errorf("output dir: %w", err)
	}
	outPath := filepath.Join(outDir, name+".mp3")
	args := []string{"-y", "-i", audioFile}
	if thumbFile != "" {
		args = append(args, "-i", thumbFile)
	}
	args = append(args, "-map", "0:a")
	if thumbFile != "" {
		args = append(args, "-map", "1")
	}
	args = append(args, "-c:a", "libmp3lame", "-b:a", "320k")
	if thumbFile != "" {
		args = append(args, "-c:v", "copy", "-id3v2_version", "3")
	}
	args = append(args, outPath)

	if err := runFFMPEG(ext.FFMPEG, args, progress); err != nil {
		return fmt.Errorf("ffmpeg conversion: %w", err)
	}
		progress(100, "Completed: "+name+".mp3")
	return nil
}

// runYTDLPJSON runs yt-dlp --dump-json and returns the first JSON object line.
func runYTDLPJSON(bin, url string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "--dump-json", "--no-warnings", "--no-playlist", "--", url)
	var out, serr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &serr
	_ = cmd.Run()
	if s := firstJSONLine(out.String()); s != "" {
		return s, nil
	}
	return "", fmt.Errorf("yt-dlp json: %s", strings.TrimSpace(serr.String()))
}

// runYTDLP runs yt-dlp, forwarding [download] percentage progress.
func runYTDLP(bin string, args []string, progress Progress) error {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	scanYTDLPProgress(stderr, progress)
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

// runFFMPEG runs ffmpeg, forwarding conversion progress (duration-based).
func runFFMPEG(bin string, args []string, progress Progress) error {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, args...)
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	scanFFMPEGProgress(stderr, progress)
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}

var (
	dlPctRE = regexp.MustCompile(`\[download\]\s+([0-9]+(?:\.[0-9]+)?)%`)
	durRE   = regexp.MustCompile(`Duration:\s+(\d+):(\d+):(\d+)`)
	timeRE  = regexp.MustCompile(`time=(\d+):(\d+):(\d+)`)
)

func scanYTDLPProgress(r io.Reader, progress Progress) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if m := dlPctRE.FindStringSubmatch(line); m != nil {
			if p, err := strconv.ParseFloat(m[1], 64); err == nil {
				progress(int(p), "Downloading...")
			}
		}
	}
}

func scanFFMPEGProgress(r io.Reader, progress Progress) {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	var total float64
	for sc.Scan() {
		line := sc.Text()
		if total == 0 {
			if m := durRE.FindStringSubmatch(line); m != nil {
				total = hmsToSec(m)
			}
		}
		if total > 0 {
			if m := timeRE.FindStringSubmatch(line); m != nil {
				cur := hmsToSec(m)
				p := int(cur / total * 100)
				if p > 100 {
					p = 100
				}
				progress(p, "Converting to MP3...")
			}
		}
	}
}

func hmsToSec(m []string) float64 {
	h, _ := strconv.Atoi(m[1])
	mi, _ := strconv.Atoi(m[2])
	s, _ := strconv.Atoi(m[3])
	return float64(h*3600 + mi*60 + s)
}

func firstJSONLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") {
			return line
		}
	}
	return ""
}

func findAudioFile(dir, base string) (string, error) {
	matches, _ := filepath.Glob(filepath.Join(dir, base+".*"))
	for _, m := range matches {
		ext := strings.ToLower(filepath.Ext(m))
		switch ext {
		case ".webm", ".m4a", ".opus", ".ogg", ".mp4", ".flac", ".mp3":
			return m, nil
		}
	}
	return "", fmt.Errorf("yt-dlp produced no audio file")
}

func findThumbnail(base string) string {
	for _, ext := range []string{"jpg", "png", "webp"} {
		p := base + "." + ext
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// ensureCoverFormat guarantees the cover is a format the mp3 muxer can write as
// an ID3v2 APIC frame. The mp3 muxer only accepts JPEG/PNG; WebP is rejected.
// Since YouTube commonly emits .webp thumbnails, we transcode those to JPEG
// with the bundled ffmpeg. Any other format is returned unchanged.
func ensureCoverFormat(ffmpeg, thumb string) string {
	if strings.ToLower(filepath.Ext(thumb)) != ".webp" {
		return thumb
	}
	jpg := strings.TrimSuffix(thumb, filepath.Ext(thumb)) + ".jpg"
	args := []string{"-y", "-i", thumb, "-frames:v", "1", jpg}
	if err := runFFMPEG(ffmpeg, args, noopProgress); err != nil {
		return thumb
	}
	return jpg
}

// noopProgress swallows progress events. Used for background ffmpeg calls that
// carry no user-facing progress (e.g. cover transcode).
func noopProgress(int, string) {}

func sanitizeFilename(name string) string {
	name = strings.Map(func(r rune) rune {
		switch r {
		case '/', '\\', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		}
		return r
	}, name)
	name = strings.TrimSpace(name)
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}
