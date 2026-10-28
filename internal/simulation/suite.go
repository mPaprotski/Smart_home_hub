// Package simulation contains only synthetic adapters. It never opens host devices.
package simulation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mPaprotski/Smart_home_hub/internal/devices"
	"github.com/mPaprotski/Smart_home_hub/internal/events"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"math/rand"
	"sync"
	"time"
)

type Suite struct {
	mu            sync.Mutex
	bus           *events.Bus
	rng           *rand.Rand
	tick          time.Duration
	epoch         string
	sequence      uint64
	state         string
	cancel        context.CancelFunc
	done          chan struct{}
	fatal         chan<- error
	Strip, Matrix *devices.MemoryLED
}

func New(bus *events.Bus, seed int64, tick time.Duration, epoch string, fatal chan<- error) *Suite {
	strip, _ := devices.NewMemoryLED(4)
	matrix, _ := devices.NewMemoryLED(4)
	return &Suite{bus: bus, rng: rand.New(rand.NewSource(seed)), tick: tick, epoch: epoch, state: "stopped", fatal: fatal, Strip: strip, Matrix: matrix}
}
func (s *Suite) Status() devices.Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return devices.Status{Kind: "synthetic", State: s.state, Simulated: true}
}

// SetState injects a sensor failure without touching a physical device.
func (s *Suite) SetState(state string) error {
	if state != "running" && state != "stale" && state != "unavailable" {
		return errors.New("invalid fake state")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
	return nil
}
func (s *Suite) Step(now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.state == "stopped" {
		return errors.New("fake adapters are stopped")
	}
	now = now.UTC()
	quality := "valid"
	if s.state != "running" {
		quality = s.state
	}
	payloads := []struct {
		kind, resource string
		payload        any
	}{
		{"SENSOR_UPDATED", "sensor.ambient", devices.Measurement{Value: 100 + s.rng.Float64()*10, Unit: "lx", Quality: quality, ObservedAt: now}},
		{"AUDIO_FEATURES_UPDATED", "audio.main", devices.AudioFeatures{RMS: s.rng.Float64(), Spectrum: [3]float64{0.2, 0.4, 0.1}, Beat: s.sequence%2 == 0, BPM: 120, Quality: quality}},
		{"PRESENCE_STATE_CHANGED", "presence.room", devices.Presence{State: []string{"PRESENT", "PROBABLY_PRESENT", "AWAY", "UNKNOWN"}[(s.sequence/5)%4], Quality: quality}},
		{"SCREEN_COLORS_OBSERVED", "screen.main", devices.ScreenColors{Epoch: s.epoch, Sequence: s.sequence + 4, Colors: []devices.RGB{{R: 32, G: 16, B: 8}}, Quality: quality, Error: fakeError(quality)}},
		{"MQTT_DEVICE_OBSERVED", "mqtt.fake.temperature", devices.MQTTObservation{Phase: mqttPhase(s.sequence, quality), DeviceID: "fake.temperature", Value: 21.5, Unit: "°C", Quality: quality}},
	}
	for _, p := range payloads {
		s.sequence++
		data, err := json.Marshal(p.payload)
		if err != nil {
			return err
		}
		e := events.Envelope{ID: fmt.Sprintf("%s-%d", s.epoch, s.sequence), Version: 1, Type: p.kind, Producer: "simulation", OccurredAt: now, ReceivedAt: now, BootEpoch: s.epoch, Sequence: s.sequence, Resource: events.Resource{Kind: resourceKind(p.kind), ID: p.resource}, Revision: s.sequence, Payload: data}
		if err = s.bus.Publish(e); err != nil {
			return err
		}
	}
	return nil
}
func (s *Suite) Start(ctx context.Context) error {
	s.mu.Lock()
	if s.done != nil || s.tick < time.Millisecond || s.epoch == "" {
		s.mu.Unlock()
		return errors.New("invalid or reused simulation")
	}
	child, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.done = make(chan struct{})
	s.state = "running"
	s.mu.Unlock()
	go func() {
		defer close(s.done)
		ticker := time.NewTicker(s.tick)
		defer ticker.Stop()
		for {
			select {
			case <-child.Done():
				return
			case now := <-ticker.C:
				if err := s.Step(now); err != nil {
					select {
					case s.fatal <- err:
					case <-child.Done():
					}
					return
				}
			}
		}
	}()
	return nil
}
func (s *Suite) Stop(ctx context.Context) error {
	s.mu.Lock()
	cancel, done := s.cancel, s.done
	s.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
		s.mu.Lock()
		s.state = "stopped"
		s.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Observer proves that observations cross the real bus before reaching Core state.
// State is bounded to the five known simulation resources.
type Observer struct {
	queue int
	mu    sync.Mutex
	bus   *events.Bus
	sub   *events.Subscription
	done  chan struct{}
	count uint64
	last  storage.Observations
	fatal chan<- error
}

func NewObserver(bus *events.Bus, store storage.Observations, queue int, fatal chan<- error) *Observer {
	return &Observer{bus: bus, last: store, queue: queue, fatal: fatal}
}
func (o *Observer) Start(ctx context.Context) error {
	if o.sub != nil {
		return errors.New("observer reused")
	}
	sub, err := o.bus.Subscribe("", o.queue, events.Control)
	if err != nil {
		return err
	}
	o.sub = sub
	o.done = make(chan struct{})
	go func() {
		defer close(o.done)
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-sub.Events():
				if !ok {
					if err := sub.Err(); err != nil && !errors.Is(err, events.ErrClosed) {
						select {
						case o.fatal <- err:
						case <-ctx.Done():
						}
					}
					return
				}
				o.mu.Lock()
				o.count++
				switch e.Resource.ID {
				case "sensor.ambient", "audio.main", "presence.room", "screen.main", "mqtt.fake.temperature":
					if err := o.last.Put(e); err != nil {
						o.mu.Unlock()
						select {
						case o.fatal <- err:
						case <-ctx.Done():
						}
						return
					}
				}
				o.mu.Unlock()
			}
		}
	}()
	return nil
}
func (o *Observer) Count() uint64 { o.mu.Lock(); defer o.mu.Unlock(); return o.count }
func (o *Observer) Last(resource string) (events.Envelope, bool) {
	o.mu.Lock()
	defer o.mu.Unlock()
	e, ok := o.last.Get(resource)
	e.Payload = append(json.RawMessage(nil), e.Payload...)
	return e, ok
}
func (o *Observer) Stop(ctx context.Context) error {
	if o.sub == nil {
		return nil
	}
	o.sub.Close()
	select {
	case <-o.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func fakeError(quality string) string {
	if quality == "unavailable" {
		return "SYNTHETIC_OFFLINE"
	}
	return ""
}
func mqttPhase(sequence uint64, quality string) string {
	if quality == "unavailable" {
		return "disconnected"
	}
	if sequence == 0 {
		return "discovery"
	}
	return "state"
}

func resourceKind(kind string) string {
	switch kind {
	case "SENSOR_UPDATED":
		return "sensor"
	case "AUDIO_FEATURES_UPDATED":
		return "audio"
	case "PRESENCE_STATE_CHANGED":
		return "presence"
	case "SCREEN_COLORS_OBSERVED":
		return "screen"
	default:
		return "device"
	}
}
