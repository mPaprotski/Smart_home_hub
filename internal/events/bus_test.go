package events

import (
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func event() Envelope {
	return Envelope{ID: "event-1", Version: 1, Type: "SENSOR_UPDATED", Producer: "test", OccurredAt: time.Unix(1, 0).UTC(), ReceivedAt: time.Unix(1, 0).UTC(), BootEpoch: "boot", Sequence: 1, Resource: Resource{Kind: "sensor", ID: "sensor.test"}, Revision: 1, Payload: json.RawMessage(`{"value":1}`)}
}
func TestIsolationAndLatest(t *testing.T) {
	b := New(10, 1024)
	defer b.Close()
	a, _ := b.Subscribe("sensor:sensor.test", 1, Latest)
	c, _ := b.Subscribe("", 2, Trigger)
	e := event()
	if err := b.Publish(e); err != nil {
		t.Fatal(err)
	}
	e.Payload[9] = '9'
	first := <-a.Events()
	first.Payload[9] = '8'
	if string((<-c.Events()).Payload) != `{"value":1}` {
		t.Fatal("payload alias")
	}
	for i := 0; i < 10; i++ {
		e = event()
		e.Sequence = uint64(i + 1)
		if err := b.Publish(e); err != nil {
			t.Fatal(err)
		}
	}
	if (<-a.Events()).Sequence != 10 {
		t.Fatal("latest value lost")
	}
	if b.Stats().Dropped == 0 {
		t.Fatal("drops invisible")
	}
}
func TestOverflowAndLimits(t *testing.T) {
	b := New(2, 128)
	a, _ := b.Subscribe("", 1, Control)
	_, err := b.Subscribe("", 1, Latest)
	if err == nil {
		t.Fatal("latest must be resource scoped")
	}
	for i := 0; i < 2; i++ {
		if err = b.Publish(event()); err != nil {
			t.Fatal(err)
		}
	}
	if !errors.Is(a.Err(), ErrOverflow) {
		t.Fatal(a.Err())
	}
	select {
	case <-a.Done():
	default:
		t.Fatal("subscriber not closed")
	}
	e := event()
	e.Payload = json.RawMessage(`{"raw_frame":"secret"}`)
	if b.Publish(e) == nil || b.Stats().Rejected != 1 {
		t.Fatal("raw frame accepted")
	}
	b.Close()
	a.Close()
	b.Close()
	if !errors.Is(b.Publish(event()), ErrClosed) {
		t.Fatal("publish after close")
	}
}
func TestConcurrentPublishUnsubscribeClose(t *testing.T) {
	b := New(128, 1024)
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				s, err := b.Subscribe("", 1, Trigger)
				if err == nil {
					_ = b.Publish(event())
					s.Close()
				}
			}
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			_ = b.Publish(event())
		}
		b.Close()
	}()
	wg.Wait()
	if b.Stats().Subscribers != 0 {
		t.Fatal("subscription leaked")
	}
}
func BenchmarkPublish(b *testing.B) {
	for _, n := range []int{0, 1, 10, 100} {
		b.Run(stringID(n), func(b *testing.B) {
			bus := New(128, 1024)
			defer bus.Close()
			for i := 0; i < n; i++ {
				_, _ = bus.Subscribe("sensor:sensor.test", 1, Latest)
			}
			e := event()
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = bus.Publish(e)
			}
		})
	}
}
func stringID(n int) string {
	if n == 0 {
		return "none"
	}
	if n == 1 {
		return "one_slow"
	}
	if n == 10 {
		return "ten_slow"
	}
	return "hundred_slow"
}

func TestFilterOrderAndFastConsumer(t *testing.T) {
	b := New(4, 1024)
	defer b.Close()
	slow, _ := b.Subscribe("sensor:sensor.test", 1, Control)
	fast, _ := b.Subscribe("sensor:sensor.test", 128, Trigger)
	other, _ := b.Subscribe("sensor:sensor.other", 1, Trigger)
	for i := 1; i <= 100; i++ {
		e := event()
		e.Sequence = uint64(i)
		if err := b.Publish(e); err != nil {
			t.Fatal(err)
		}
	}
	for i := 1; i <= 100; i++ {
		if e := <-fast.Events(); e.Sequence != uint64(i) {
			t.Fatal("per publisher order lost")
		}
	}
	if len(other.Events()) != 0 {
		t.Fatal("filter mismatch")
	}
	if !errors.Is(slow.Err(), ErrOverflow) {
		t.Fatal("slow control did not resync")
	}
	if b.Stats().Subscribers != 2 {
		t.Fatal(b.Stats())
	}
}

func TestResourceKeyBounds(t *testing.T) {
	b := New(2, 1024)
	defer b.Close()
	e := event()
	e.Resource.ID = strings.Repeat("x", 128)
	sub, err := b.Subscribe(e.Resource.Key(), 1, Latest)
	if err != nil {
		t.Fatal(err)
	}
	if err = b.Publish(e); err != nil {
		t.Fatal(err)
	}
	if got := <-sub.Events(); got.Resource != e.Resource {
		t.Fatal("long resource filter mismatch")
	}
	e.Resource.Kind = "sensor:other"
	if b.Publish(e) == nil {
		t.Fatal("ambiguous resource key accepted")
	}
}
