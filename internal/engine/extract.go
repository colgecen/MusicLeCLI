package engine

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

// Extracted holds the on-disk paths to the fully extracted engine tools. It is
// the value returned by Extract() and is what the wrappers operate on.
type Extracted struct {
	Root   string // cache root, e.g. ~/.cache/musicle/engine-v2
	YTDLP  string // absolute path to the embedded yt-dlp binary
	FFMPEG string // absolute path to the embedded ffmpeg binary
}

// cacheVersion names the on-disk engine cache. Bump it EVERY time anything
// under engine_bin/ changes (yt-dlp/ffmpeg/wrapper); otherwise installs keep
// running the previously extracted (stale) tools forever, since the sentinel
// check below cannot see content changes.
const cacheVersion = "engine-v3"

// sentinelName marks a successful extract; its presence short-circuits Extract.
const sentinelName = ".extracted"

var (
	extractOnce sync.Once
	extracted   *Extracted
	extractErr  error
)

// Extract returns the on-disk locations of the embedded tools, extracting them
// to UserCacheDir on first call. Subsequent calls return the cached result.
//
// If the embedded tools are absent or corrupt, Extract returns
// ErrNoEmbeddedAssets.
func Extract() (*Extracted, error) {
	extractOnce.Do(func() {
		extracted, extractErr = doExtract()
	})
	return extracted, extractErr
}

// MustExtract is like Extract but panics on error. Intended for use after a
// successful startup probe.
func MustExtract() *Extracted {
	e, err := Extract()
	if err != nil {
		panic(err)
	}
	return e
}

func doExtract() (*Extracted, error) {
	// Probe the embedded FS variables. Without the build tag the stub
	// implementation returns ErrNoEmbeddedAssets.
	if err := probeEmbedded(); err != nil {
		return nil, err
	}

	cacheRoot, err := cacheDir()
	if err != nil {
		return nil, err
	}

	root := filepath.Join(cacheRoot, cacheVersion)

	// Fast path: already extracted.
	if valid, _ := isExtracted(root); valid {
		return buildExtracted(root)
	}

	// Slow path: wipe and re-extract.
	if err := os.RemoveAll(root); err != nil {
		return nil, fmt.Errorf("engine: cannot clear old cache: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("engine: cannot create cache dir: %w", err)
	}

	binDst := filepath.Join(root, "engine_bin")
	if err := copyFS(binFS, "engine_bin", binDst); err != nil {
		return nil, fmt.Errorf("engine: cannot extract tools: %w", err)
	}

	// chmod +x on the tool binaries on unix. embed.FS reports every file
	// as 0444, so copyFS alone leaves them non-executable — every binary
	// the wrapper shells out to (including yt-dlp.real) must be listed here.
	if runtime.GOOS != "windows" {
		for _, name := range []string{"yt-dlp", "yt-dlp.real", "ffmpeg"} {
			_ = os.Chmod(filepath.Join(binDst, name+exeSuffix()), 0o755)
		}
	}

	// Stamp the cache as valid.
	if err := os.WriteFile(filepath.Join(root, sentinelName), []byte(cacheVersion+"\n"), 0o644); err != nil {
		return nil, fmt.Errorf("engine: cannot write sentinel: %w", err)
	}

	return buildExtracted(root)
}

func cacheDir() (string, error) {
	if d := os.Getenv("MUSICLE_CACHE_DIR"); d != "" {
		return d, nil
	}
	base, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("engine: user cache dir not found: %w", err)
	}
	return filepath.Join(base, "musicle"), nil
}

func isExtracted(root string) (bool, error) {
	data, err := os.ReadFile(filepath.Join(root, sentinelName))
	if err != nil {
		return false, err
	}
	return string(data) == cacheVersion+"\n", nil
}

func buildExtracted(root string) (*Extracted, error) {
	bin := filepath.Join(root, "engine_bin")
	ytdlp := filepath.Join(bin, "yt-dlp"+exeSuffix())
	ffmpeg := filepath.Join(bin, "ffmpeg"+exeSuffix())

	for _, p := range []string{ytdlp, ffmpeg} {
		if _, err := os.Stat(p); err != nil {
			return nil, fmt.Errorf("engine: expected binary missing: %s (%w)", p, err)
		}
	}

	return &Extracted{
		Root:   root,
		YTDLP:  ytdlp,
		FFMPEG: ffmpeg,
	}, nil
}

func exeSuffix() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// copyFS recursively copies an embed.FS subtree rooted at srcRoot into dst.
// Permissions follow fs.FileInfo when available; directories are 0o755,
// regular files are 0o644 (the caller chmods executables afterwards).
func copyFS(files fs.FS, srcRoot, dst string) error {
	return fs.WalkDir(files, srcRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if path == srcRoot {
			return nil
		}
		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		target := filepath.Join(dst, filepath.FromSlash(rel))

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := files.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		info, err := d.Info()
		if err != nil {
			info = nil
		}
		mode := os.FileMode(0o644)
		if info != nil && info.Mode()&0o111 != 0 {
			mode = 0o755
		}

		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, in); err != nil {
			out.Close()
			return err
		}
		return out.Close()
	})
}
