// Feasibility spike only: logical zones, mapping and power limits. No GPIO I/O.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
)

type RGB [3]uint8
type Zone struct {
	ID            string
	Offset, Count int
	Limit         float64
}

func validate(zones []Zone, capacity int) error {
	if capacity <= 0 || capacity > 4096 {
		return fmt.Errorf("capacity must be positive")
	}
	ids := map[string]bool{}
	used := make([]bool, capacity)
	for _, z := range zones {
		if z.ID == "" || ids[z.ID] || z.Offset < 0 || z.Count <= 0 || z.Offset > capacity || z.Count > capacity-z.Offset || math.IsNaN(z.Limit) || math.IsInf(z.Limit, 0) || z.Limit < 0 {
			return fmt.Errorf("invalid zone")
		}
		ids[z.ID] = true
		for i := z.Offset; i < z.Offset+z.Count; i++ {
			if used[i] {
				return fmt.Errorf("overlap")
			}
			used[i] = true
		}
	}
	return nil
}
func current(frame []RGB, idle, full float64) float64 {
	amps := float64(len(frame)) * idle
	for _, p := range frame {
		for _, c := range p {
			amps += (full - idle) * float64(c) / (3 * 255)
		}
	}
	return amps
}
func limit(frame []RGB, idle, full, max float64) ([]RGB, error) {
	if idle < 0 || full < idle || max < 0 || math.IsNaN(idle) || math.IsInf(idle, 0) || math.IsNaN(full) || math.IsInf(full, 0) || math.IsNaN(max) || math.IsInf(max, 0) {
		return nil, fmt.Errorf("bad current model")
	}
	base := float64(len(frame)) * idle
	if base > max+1e-12 {
		return nil, fmt.Errorf("limit below idle floor")
	}
	amps := current(frame, idle, full)
	scale := 1.0
	if amps > max && amps > base {
		scale = (max - base) / (amps - base)
	}
	out := make([]RGB, len(frame))
	for i, p := range frame {
		for j, c := range p {
			out[i][j] = uint8(math.Floor(float64(c) * scale))
		}
	}
	return out, nil
}
func run() error {
	zones := []Zone{{"monitor.backlight", 0, 120, 4}, {"desk.wall", 120, 120, 4}}
	if err := validate(zones, 240); err != nil {
		return err
	}
	frame := make([]RGB, 240)
	for i := 0; i < 120; i++ {
		frame[i] = RGB{255, 0, 0}
	}
	for i := 120; i < 240; i++ {
		frame[i] = RGB{0, 0, 255}
	}
	for i := 0; i < 120; i++ {
		frame[i] = RGB{}
	}
	for i := 120; i < 240; i++ {
		if frame[i] != (RGB{0, 0, 255}) {
			return fmt.Errorf("zone isolation failed")
		}
	}
	bad := [][]Zone{
		{{"a", 0, 121, 1}, {"b", 120, 120, 1}},
		{{"a", -1, 120, 1}}, {{"a", 0, 0, 1}}, {{"a", 0, 241, 1}},
		{{"a", 0, 120, 1}, {"a", 120, 120, 1}}, {{"a", 0, 120, math.NaN()}},
	}
	for _, v := range bad {
		if validate(v, 240) == nil {
			return fmt.Errorf("bad mapping accepted")
		}
	}
	for i := range frame {
		frame[i] = RGB{255, 255, 255}
	}
	for _, z := range zones {
		p, e := limit(frame[z.Offset:z.Offset+z.Count], .001, .06, z.Limit)
		if e != nil {
			return e
		}
		copy(frame[z.Offset:], p)
	}
	out, e := limit(frame, .001, .06, 6)
	if e != nil {
		return e
	}
	if current(out, .001, .06) > 6+1e-9 {
		return fmt.Errorf("global limiter failed")
	}
	if _, e := limit(frame, .001, .06, .1); e == nil {
		return fmt.Errorf("idle floor accepted")
	}
	zero, e := limit(make([]RGB, 240), .001, .06, .24)
	if e != nil || current(zero, .001, .06) > .240000001 {
		return fmt.Errorf("zero frame failed")
	}
	result := map[string]interface{}{"probe": "D04-zone-frame", "status": "PASS-SIM", "checks": []string{"two independent slices", "six invalid mappings rejected", "zone limits then global 6A", "idle floor rejected", "zero frame"}, "synthetic_led_count": 240, "limited_current_a": current(out, .001, .06), "hardware_io": false, "language": "Go"}
	return json.NewEncoder(os.Stdout).Encode(result)
}
func main() {
	if e := run(); e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}
