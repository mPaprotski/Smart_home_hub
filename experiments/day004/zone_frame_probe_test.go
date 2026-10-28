package main

import (
	"math"
	"math/rand"
	"testing"
)

func TestLimiterRespectsBudgetAndDoesNotRaiseChannels(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	for trial := 0; trial < 1000; trial++ {
		count := 1 + rng.Intn(250)
		frame := make([]RGB, count)
		for i := range frame {
			frame[i] = RGB{uint8(rng.Intn(256)), uint8(rng.Intn(256)), uint8(rng.Intn(256))}
		}
		idle := .001
		full := .06
		budget := float64(count)*idle + rng.Float64()*float64(count)*(full-idle)
		out, err := limit(frame, idle, full, budget)
		if err != nil {
			t.Fatal(err)
		}
		if current(out, idle, full) > budget+1e-9 {
			t.Fatalf("trial %d over budget", trial)
		}
		for i, p := range out {
			for c, v := range p {
				if v > frame[i][c] {
					t.Fatal("limiter brightened a channel")
				}
			}
		}
	}
}
func TestInvalidCurrentModelsRejected(t *testing.T) {
	cases := []struct{ idle, full, max float64 }{{math.NaN(), .06, 1}, {.001, math.NaN(), 1}, {.001, .06, math.NaN()}, {.001, math.Inf(1), 1}, {.001, .06, -1}, {.01, .001, 1}, {.001, .06, 0}}
	for _, v := range cases {
		if _, err := limit([]RGB{{255, 255, 255}}, v.idle, v.full, v.max); err == nil {
			t.Fatalf("accepted invalid model %#v", v)
		}
	}
}
func TestCapacityAndAdjacentSlices(t *testing.T) {
	if validate([]Zone{{"a", 0, 1, 1}, {"b", 1, 1, 1}}, 2) != nil {
		t.Fatal("adjacent slices rejected")
	}
	if validate([]Zone{{"a", 0, 2, 1}, {"b", 1, 1, 1}}, 2) == nil {
		t.Fatal("overlap accepted")
	}
	if validate([]Zone{{"a", 0, 1, 1}}, 1000000000) == nil {
		t.Fatal("unbounded allocation accepted")
	}
}
