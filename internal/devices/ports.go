package devices

import (
	"context"
	"errors"
	"math"
	"sync"
	"time"
)

type Status struct {
	Kind      string `json:"kind"`
	State     string `json:"state"`
	Simulated bool   `json:"simulated"`
}
type Adapter interface {
	Start(context.Context) error
	Stop(context.Context) error
	Status() Status
}
type RGB struct {
	R uint8 `json:"r"`
	G uint8 `json:"g"`
	B uint8 `json:"b"`
}
type Measurement struct {
	Value      float64   `json:"value"`
	Unit       string    `json:"unit"`
	Quality    string    `json:"quality"`
	ObservedAt time.Time `json:"observed_at"`
}

func (m Measurement) Validate() error {
	if math.IsNaN(m.Value) || math.IsInf(m.Value, 0) || m.Unit == "" || m.ObservedAt.IsZero() || (m.Quality != "valid" && m.Quality != "stale" && m.Quality != "unavailable") {
		return errors.New("invalid measurement")
	}
	return nil
}

// LED accepts full frames only; the fake keeps one owned copy, not frame history.
type LED interface {
	Show([]RGB) error
	Frame() []RGB
}
type MemoryLED struct {
	mu     sync.Mutex
	pixels []RGB
	frames uint64
}

func NewMemoryLED(count int) (*MemoryLED, error) {
	if count < 1 || count > 4096 {
		return nil, errors.New("pixel count outside bounds")
	}
	return &MemoryLED{pixels: make([]RGB, count)}, nil
}
func (l *MemoryLED) Show(frame []RGB) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if len(frame) != len(l.pixels) {
		return errors.New("frame size mismatch")
	}
	copy(l.pixels, frame)
	l.frames++
	return nil
}
func (l *MemoryLED) Frame() []RGB {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]RGB(nil), l.pixels...)
}
func MatrixIndex(x, y, width, height int) (int, error) {
	if width < 1 || height < 1 || width > 4096 || height > 4096 || width*height > 4096 || x < 0 || y < 0 || x >= width || y >= height {
		return 0, errors.New("matrix coordinate outside bounds")
	}
	if y%2 == 1 {
		x = width - 1 - x
	}
	return y*width + x, nil
}

func (l *MemoryLED) Frames() uint64 { l.mu.Lock(); defer l.mu.Unlock(); return l.frames }

// Mapping transforms logical coordinates clockwise before optional physical flips
// and serpentine wiring. Rotation 90/270 swaps physical width and height.
type Mapping struct {
	Width, Height, Rotation  int
	FlipX, FlipY, Serpentine bool
}

func (m Mapping) Index(x, y int) (int, error) {
	if m.Width < 1 || m.Height < 1 || m.Width > 4096 || m.Height > 4096 || m.Width*m.Height > 4096 || x < 0 || y < 0 || x >= m.Width || y >= m.Height {
		return 0, errors.New("invalid matrix mapping")
	}
	w, h := m.Width, m.Height
	switch m.Rotation {
	case 0:
	case 90:
		x, y, w, h = h-1-y, x, h, w
	case 180:
		x, y = w-1-x, h-1-y
	case 270:
		x, y, w, h = y, w-1-x, h, w
	default:
		return 0, errors.New("rotation must be 0, 90, 180 or 270")
	}
	if m.FlipX {
		x = w - 1 - x
	}
	if m.FlipY {
		y = h - 1 - y
	}
	if m.Serpentine && y%2 == 1 {
		x = w - 1 - x
	}
	return y*w + x, nil
}
