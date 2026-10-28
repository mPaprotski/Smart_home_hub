package simulation

import (
	"context"
	"encoding/json"
	"github.com/mPaprotski/Smart_home_hub/internal/devices"
	"github.com/mPaprotski/Smart_home_hub/internal/events"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"reflect"
	"testing"
	"time"
)

func TestDeterministicObservations(t *testing.T) {
	makeEvents := func() []events.Envelope {
		b := events.New(8, 4096)
		defer b.Close()
		sub, _ := b.Subscribe("", 8, events.Control)
		s := New(b, 42, time.Hour, "boot", nil)
		if err := s.SetState("running"); err != nil {
			t.Fatal(err)
		}
		if err := s.Step(time.Unix(100, 0)); err != nil {
			t.Fatal(err)
		}
		result := []events.Envelope{}
		for i := 0; i < 5; i++ {
			result = append(result, <-sub.Events())
		}
		return result
	}
	if !reflect.DeepEqual(makeEvents(), makeEvents()) {
		t.Fatal("seed and clock not deterministic")
	}
}
func TestSensorThroughCoreBus(t *testing.T) {
	b := events.New(8, 4096)
	defer b.Close()
	fatal := make(chan error, 1)
	o := NewObserver(b, storage.NewMemory(5), 64, fatal)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := o.Start(ctx); err != nil {
		t.Fatal(err)
	}
	s := New(b, 42, time.Millisecond, "integration", fatal)
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for o.Count() < 5 && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	e, ok := o.Last("sensor:sensor.ambient")
	if !ok || e.Type != "SENSOR_UPDATED" || e.Producer != "simulation" {
		t.Fatal("sensor did not reach core", e)
	}
	if err := s.SetState("stale"); err != nil {
		t.Fatal(err)
	}
	stopCtx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := s.Stop(stopCtx); err != nil {
		t.Fatal(err)
	}
	if err := o.Stop(stopCtx); err != nil {
		t.Fatal(err)
	}
	if b.Stats().Subscribers != 0 {
		t.Fatal("leaked subscription")
	}
}

func TestAllFakeScenarios(t *testing.T) {
	b := events.New(8, 4096)
	defer b.Close()
	sub, _ := b.Subscribe("", 64, events.Control)
	s := New(b, 42, time.Hour, "boot", nil)
	if err := s.SetState("running"); err != nil {
		t.Fatal(err)
	}
	states := map[string]bool{}
	for i := 0; i < 4; i++ {
		if err := s.Step(time.Unix(int64(i+1), 0)); err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 5; j++ {
			e := <-sub.Events()
			switch e.Type {
			case "PRESENCE_STATE_CHANGED":
				var p devices.Presence
				if err := json.Unmarshal(e.Payload, &p); err != nil {
					t.Fatal(err)
				}
				states[p.State] = true
			case "AUDIO_FEATURES_UPDATED":
				var a devices.AudioFeatures
				if err := json.Unmarshal(e.Payload, &a); err != nil || a.BPM != 120 || a.RMS < 0 || a.RMS > 1 {
					t.Fatal(a, err)
				}
			case "SCREEN_COLORS_OBSERVED":
				var c devices.ScreenColors
				if err := json.Unmarshal(e.Payload, &c); err != nil || c.Sequence != e.Sequence || c.Epoch != e.BootEpoch || len(c.Colors) != 1 {
					t.Fatal(c, err)
				}
			case "MQTT_DEVICE_OBSERVED":
				var m devices.MQTTObservation
				if err := json.Unmarshal(e.Payload, &m); err != nil {
					t.Fatal(err)
				}
				if (i == 0 && m.Phase != "discovery") || (i > 0 && m.Phase != "state") {
					t.Fatal(m)
				}
			}
		}
	}
	if len(states) != 4 {
		t.Fatal(states)
	}
	for _, state := range []string{"stale", "unavailable"} {
		if err := s.SetState(state); err != nil {
			t.Fatal(err)
		}
		if err := s.Step(time.Unix(10, 0)); err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 5; j++ {
			e := <-sub.Events()
			var p map[string]any
			_ = json.Unmarshal(e.Payload, &p)
			if p["quality"] != state {
				t.Fatal(p)
			}
			if state == "unavailable" && e.Type == "SCREEN_COLORS_OBSERVED" && p["error"] != "SYNTHETIC_OFFLINE" {
				t.Fatal(p)
			}
			if state == "unavailable" && e.Type == "MQTT_DEVICE_OBSERVED" && p["phase"] != "disconnected" {
				t.Fatal(p)
			}
		}
	}
}
