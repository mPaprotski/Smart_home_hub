package system

import (
	"github.com/mPaprotski/Smart_home_hub/internal/config"
	"testing"
)

func TestCriticalOptionalAndSnapshots(t *testing.T) {
	c := config.Defaults()
	h := New(c)
	if h.Snapshot().Ready {
		t.Fatal("starting adapters ready")
	}
	h.Set("adapters", "READY")
	h.SetPhase("RUNNING")
	if !h.Snapshot().Ready {
		t.Fatal("simulation not ready")
	}
	for _, state := range []string{"UNAVAILABLE", "STALE", "NOT_IMPLEMENTED"} {
		h.Set("ml", state)
		s := h.Snapshot()
		if !s.Ready || !s.Degraded {
			t.Fatal(state, s)
		}
		h.Set("storage", state)
		if h.Snapshot().Ready {
			t.Fatal("critical ready", state)
		}
		h.Set("storage", "READY")
	}
	h.Set("ml", "DISABLED")
	s := h.Snapshot()
	if s.Degraded || s.Persistent || !s.Simulated {
		t.Fatal(s)
	}
	s.Dependencies[0].Status = "UNAVAILABLE"
	if !h.Snapshot().Ready {
		t.Fatal("snapshot alias")
	}
	h.SetPhase("SAFE_MODE")
	if h.Snapshot().Ready || !h.Snapshot().Live {
		t.Fatal("safe mode readiness")
	}
	h.Stopping()
	if h.Snapshot().Ready || h.Snapshot().Live {
		t.Fatal("stopping live")
	}
}

func TestPersistentSimulationAndRequiredAuth(t *testing.T) {
	c := config.Defaults()
	c.Storage = "sqlite"
	c.Auth.Required = true
	h := New(c)
	h.Set("adapters", "READY")
	h.SetPhase("RUNNING")
	if h.Snapshot().Persistent || h.Snapshot().Ready {
		t.Fatal("uninitialized SQLite ready")
	}
	h.Set("storage", "READY")
	s := h.Snapshot()
	if !s.Persistent || !s.Simulated || s.Ready {
		t.Fatal("missing sessions declared ready", s)
	}
	h.Set("storage", "DISABLED")
	if h.Snapshot().Ready || h.Snapshot().Persistent {
		t.Fatal("disabled critical counted as ready")
	}
	c.Auth.Required = false
	h = New(c)
	h.Set("adapters", "READY")
	h.Set("storage", "READY")
	h.SetPhase("RUNNING")
	if !h.Snapshot().Ready || !h.Snapshot().Persistent {
		t.Fatal("explicit development profile unavailable")
	}
}
