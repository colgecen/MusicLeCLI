# Makefile — reproducible build, vet and test for musicle-cli
#
# The full binary needs CGO + ALSA dev headers (the oto audio backend). On a
# machine without those, `make build` fails only on the player package; the
# engine and download packages still build/vet/test via `make check`.

GO ?= go
VERSION ?= 1.5.0

LDFLAGS := -ldflags="-s -w -X main.version=$(VERSION)"

.PHONY: all engine build vet test check clean

all: build

# Build the embedded download engine (standalone yt-dlp + static ffmpeg/ffprobe).
engine:
	./scripts/prepare-engine.sh linux

# Full binary build (requires CGO + ALSA dev headers for the audio player).
# Depends on `engine` so the bundled yt-dlp+ffmpeg are prepared, then baked in
# with -tags engine_assets. The result is fully self-contained for end users.
# The binary is stripped (-s -w) and UPX-packed to keep the embedded tools small.
build: engine
	CGO_ENABLED=1 $(GO) build -tags engine_assets $(LDFLAGS) -o build/musicle-cli .
	-upx --best --lzma build/musicle-cli 2>/dev/null || true

# Vet the packages that do not depend on the audio player (buildable anywhere).
vet:
	$(GO) vet ./internal/... ./bridge/download/...

# Unit tests (engine package + download helpers).
test:
	$(GO) test ./internal/... ./bridge/download/...

# Fast local check: everything that compiles without the cgo audio backend.
check: vet test

clean:
	rm -rf build internal/engine/engine_bin
