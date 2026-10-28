package storage

import (
	"github.com/mPaprotski/Smart_home_hub/internal/events"
	"testing"
)

func TestBoundedOwnership(t *testing.T) {
	m := NewMemory(1)
	e := events.Envelope{Resource: events.Resource{Kind: "sensor", ID: "one"}, Payload: []byte(`{"n":1}`)}
	if err := m.Put(e); err != nil {
		t.Fatal(err)
	}
	e.Payload[5] = '9'
	got, _ := m.Get("sensor:one")
	if string(got.Payload) != `{"n":1}` {
		t.Fatal("input alias")
	}
	got.Payload[5] = '8'
	got, _ = m.Get("sensor:one")
	if string(got.Payload) != `{"n":1}` {
		t.Fatal("output alias")
	}
	e.Resource.ID = "two"
	if m.Put(e) == nil {
		t.Fatal("unbounded store")
	}
	if _, ok := NewMemory(1).Get("sensor:one"); ok {
		t.Fatal("data persisted across stores")
	}
}
