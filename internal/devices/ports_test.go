package devices

import (
	"math"
	"testing"
	"time"
)

func TestFrameOwnershipAndMapping(t *testing.T) {
	led, err := NewMemoryLED(4)
	if err != nil {
		t.Fatal(err)
	}
	frame := []RGB{{R: 1}, {G: 2}, {B: 3}, {R: 4}}
	if err = led.Show(frame); err != nil {
		t.Fatal(err)
	}
	frame[0].R = 99
	out := led.Frame()
	out[1].G = 99
	if led.Frame()[0].R != 1 || led.Frame()[1].G != 2 {
		t.Fatal("frame alias")
	}
	if led.Show(frame[:3]) == nil {
		t.Fatal("short frame")
	}
	for _, c := range []struct{ x, y, want int }{{0, 0, 0}, {1, 0, 1}, {0, 1, 3}, {1, 1, 2}} {
		got, err := MatrixIndex(c.x, c.y, 2, 2)
		if err != nil || got != c.want {
			t.Fatal(got, err)
		}
	}
	if _, err := MatrixIndex(-1, 0, 2, 2); err == nil {
		t.Fatal("invalid coordinate")
	}
}
func TestMeasurementFinite(t *testing.T) {
	m := Measurement{Value: math.NaN(), Unit: "lx", Quality: "valid", ObservedAt: time.Now()}
	if m.Validate() == nil {
		t.Fatal("NaN accepted")
	}
}

func TestRotatedMatrixBijection(t *testing.T) {
	for _, rotation := range []int{0, 90, 180, 270} {
		for _, flipX := range []bool{false, true} {
			m := Mapping{Width: 3, Height: 2, Rotation: rotation, FlipX: flipX, Serpentine: true}
			seen := map[int]bool{}
			for y := 0; y < 2; y++ {
				for x := 0; x < 3; x++ {
					i, err := m.Index(x, y)
					if err != nil || i < 0 || i >= 6 || seen[i] {
						t.Fatal(i, err)
					}
					seen[i] = true
				}
			}
		}
	}
	m := Mapping{Width: 2, Height: 2, Rotation: 90}
	i, err := m.Index(0, 0)
	if err != nil || i != 1 {
		t.Fatal(i, err)
	}
}
