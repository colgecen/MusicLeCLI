package bridge

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gopxl/beep"
)

// countStreamer records Seek calls; Len 0 simulates an empty stream.
type countStreamer struct {
	mu     sync.Mutex
	pos    int
	length int
	seeks  int
}

func (s *countStreamer) Stream(samples [][2]float64) (int, bool) { return 0, false }
func (s *countStreamer) Len() int                                { return s.length }
func (s *countStreamer) Position() int                           { return s.pos }
func (s *countStreamer) Close() error                            { return nil }
func (s *countStreamer) Err() error                              { return nil }
func (s *countStreamer) Seek(p int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.seeks++
	s.pos = p
	return nil
}

// panicStreamer blows up on Seek, like a corrupted decoder would.
type panicStreamer struct{ countStreamer }

func (s *panicStreamer) Seek(p int) error { panic("decoder boom") }

func savePlayer() (beep.StreamSeekCloser, beep.Format, string, bool) {
	player.mu.Lock()
	defer player.mu.Unlock()
	return player.streamer, player.format, player.currentFile, player.paused
}

func restorePlayer(st beep.StreamSeekCloser, f beep.Format, file string, paused bool) {
	player.mu.Lock()
	defer player.mu.Unlock()
	player.streamer, player.format, player.currentFile, player.paused = st, f, file, paused
	player.startTime = time.Now()
	player.pauseOffset = 0
}

// TestSeekEmptyStream: Len()==0 must error, never Seek(-1).
func TestSeekEmptyStream(t *testing.T) {
	st, f, file, paused := savePlayer()
	defer restorePlayer(st, f, file, paused)

	fake := &countStreamer{length: 0}
	player.mu.Lock()
	player.streamer = fake
	player.format = beep.Format{SampleRate: 44100}
	player.mu.Unlock()

	res := player.seek(5)
	if res.Status != "error" {
		t.Fatalf("bos akista Status=%q, hata bekleniyordu", res.Status)
	}
	fake.mu.Lock()
	defer fake.mu.Unlock()
	if fake.seeks != 0 {
		t.Fatalf("Seek cagrilmamaliydi, %d kez cagrilmis", fake.seeks)
	}
}

// TestSeekConcurrentRapid: ard arda sag/sol basimi <-/-> carpismamali.
func TestSeekConcurrentRapid(t *testing.T) {
	st, f, file, paused := savePlayer()
	defer restorePlayer(st, f, file, paused)

	fake := &countStreamer{length: 44100 * 180}
	player.mu.Lock()
	player.streamer = fake
	player.format = beep.Format{SampleRate: 44100}
	player.currentFile = "test"
	player.paused = true
	player.mu.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				player.seek(5)
			} else {
				player.seek(-5)
			}
		}(i)
	}
	wg.Wait()

	player.mu.Lock()
	off := player.pauseOffset
	n := fake.seeks
	player.mu.Unlock()
	_ = n
	if off < 0 || off > 200 {
		t.Fatalf("konum aralik disi: %v", off)
	}
}

// TestPlayerCallRecover: panik uygulamayi oldurmemeli, hata donmeli.
func TestPlayerCallRecover(t *testing.T) {
	st, f, file, paused := savePlayer()
	defer restorePlayer(st, f, file, paused)

	player.mu.Lock()
	player.streamer = &panicStreamer{countStreamer{length: 44100}}
	player.format = beep.Format{SampleRate: 44100}
	player.mu.Unlock()

	res, err := PlayerCall(Action{Action: "seek", Value: 5})
	if err != nil {
		t.Fatalf("hata donmemeli: %v", err)
	}
	if res == nil || res.Status != "error" || !strings.Contains(res.Error, "player panic") {
		t.Fatalf("panic kurtarma calismadi: %+v", res)
	}
}
