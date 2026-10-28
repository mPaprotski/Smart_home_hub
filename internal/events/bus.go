// Package events provides a bounded in-process bus. It is not a durable journal.
package events

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

var ErrClosed = errors.New("event bus closed")
var ErrOverflow = errors.New("subscriber overflow: resync required")

type Policy uint8

const (
	Control Policy = iota
	Trigger
	Latest
)

type Resource struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

func (r Resource) Key() string { return r.Kind + ":" + r.ID }

type Envelope struct {
	ID            string          `json:"event_id"`
	Version       int             `json:"schema_version"`
	Type          string          `json:"type"`
	Producer      string          `json:"producer"`
	OccurredAt    time.Time       `json:"occurred_at"`
	ReceivedAt    time.Time       `json:"received_at"`
	BootEpoch     string          `json:"boot_epoch"`
	Sequence      uint64          `json:"sequence"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	CausationID   string          `json:"causation_id,omitempty"`
	Resource      Resource        `json:"resource"`
	Revision      uint64          `json:"revision"`
	Payload       json.RawMessage `json:"payload"`
}

var catalog = map[string]bool{"SENSOR_UPDATED": true, "AUDIO_FEATURES_UPDATED": true, "PRESENCE_STATE_CHANGED": true, "SCREEN_COLORS_OBSERVED": true, "MQTT_DEVICE_OBSERVED": true, "DEVICE_CONNECTED": true, "SYSTEM_STATE_CHANGED": true}

func (e Envelope) Validate(max int) error {
	_, occurredOffset := e.OccurredAt.Zone()
	_, receivedOffset := e.ReceivedAt.Zone()
	if occurredOffset != 0 || receivedOffset != 0 {
		return errors.New("timestamps must be UTC")
	}
	if e.Version != 1 || !catalog[e.Type] || e.ID == "" || len(e.ID) > 128 || e.Producer == "" || len(e.Producer) > 128 || e.BootEpoch == "" || len(e.BootEpoch) > 128 || e.Resource.ID == "" || len(e.Resource.ID) > 128 || e.Resource.Kind == "" || len(e.Resource.Kind) > 32 || strings.Contains(e.Resource.Kind, ":") || len(e.CorrelationID) > 128 || len(e.CausationID) > 128 || e.Sequence == 0 || e.Revision == 0 || e.OccurredAt.IsZero() || e.ReceivedAt.IsZero() {
		return errors.New("invalid event envelope")
	}
	if len(e.Payload) > max || len(e.Payload) == 0 {
		return errors.New("invalid payload size")
	}
	var p map[string]json.RawMessage
	if json.Unmarshal(e.Payload, &p) != nil || p == nil {
		return errors.New("payload must be a JSON object")
	}
	// This bus carries derived observations, never source media.
	for _, key := range []string{"raw_frame", "raw_audio", "image", "audio_samples", "frame_bytes"} {
		if _, ok := p[key]; ok {
			return errors.New("raw media is forbidden")
		}
	}
	return nil
}
func clone(e Envelope) Envelope { e.Payload = append(json.RawMessage(nil), e.Payload...); return e }

type Stats struct {
	Published   uint64
	Delivered   uint64
	Dropped     uint64
	Overflows   uint64
	Rejected    uint64
	Subscribers int
}
type Bus struct {
	mu                  sync.Mutex
	closed              bool
	maxSubs, maxPayload int
	next                uint64
	subs                map[uint64]*Subscription
	stats               Stats
}
type Subscription struct {
	bus      *Bus
	id       uint64
	resource string
	policy   Policy
	ch       chan Envelope
	done     chan struct{}
	err      error
}

func New(maxSubs, maxPayload int) *Bus {
	return &Bus{maxSubs: maxSubs, maxPayload: maxPayload, subs: make(map[uint64]*Subscription)}
}
func (b *Bus) Subscribe(resource string, capacity int, policy Policy) (*Subscription, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return nil, ErrClosed
	}
	if len(b.subs) >= b.maxSubs || capacity < 1 || capacity > 4096 || len(resource) > 161 || policy > Latest {
		return nil, errors.New("invalid subscription or subscription limit")
	}
	if policy == Latest && (resource == "" || capacity != 1) {
		return nil, errors.New("latest requires one slot and an exact resource")
	}
	b.next++
	s := &Subscription{bus: b, id: b.next, resource: resource, policy: policy, ch: make(chan Envelope, capacity), done: make(chan struct{})}
	b.subs[s.id] = s
	return s, nil
}
func (s *Subscription) Events() <-chan Envelope { return s.ch }
func (s *Subscription) Done() <-chan struct{}   { return s.done }
func (s *Subscription) Err() error              { s.bus.mu.Lock(); defer s.bus.mu.Unlock(); return s.err }
func (s *Subscription) Close()                  { s.bus.mu.Lock(); defer s.bus.mu.Unlock(); s.bus.remove(s, nil) }
func (b *Bus) remove(s *Subscription, err error) {
	if _, ok := b.subs[s.id]; !ok {
		return
	}
	delete(b.subs, s.id)
	s.err = err
	close(s.ch)
	close(s.done)
}
func (b *Bus) Publish(e Envelope) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return ErrClosed
	}
	if err := e.Validate(b.maxPayload); err != nil {
		b.stats.Rejected++
		return fmt.Errorf("event rejected: %w", err)
	}
	b.stats.Published++
	for _, s := range b.subs {
		if s.resource != "" && s.resource != e.Resource.Key() {
			continue
		}
		select {
		case s.ch <- clone(e):
			b.stats.Delivered++
		default:
			switch s.policy {
			case Control:
				b.stats.Overflows++
				b.remove(s, ErrOverflow)
			case Trigger:
				b.stats.Dropped++
			case Latest:
				select {
				case <-s.ch:
					b.stats.Dropped++
				default:
				}
				s.ch <- clone(e)
				b.stats.Delivered++
			}
		}
	}
	return nil
}
func (b *Bus) Stats() Stats {
	b.mu.Lock()
	defer b.mu.Unlock()
	s := b.stats
	s.Subscribers = len(b.subs)
	return s
}
func (b *Bus) Close() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.closed {
		return
	}
	b.closed = true
	for _, s := range b.subs {
		b.remove(s, ErrClosed)
	}
}
