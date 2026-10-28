package system

import (
	"github.com/mPaprotski/Smart_home_hub/internal/config"
	"sync"
)

type Dependency struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Status   string `json:"status"`
	Critical bool   `json:"critical"`
	Reason   string `json:"reason,omitempty"`
}
type Snapshot struct {
	Phase        string       `json:"phase"`
	Live         bool         `json:"live"`
	Ready        bool         `json:"ready"`
	Degraded     bool         `json:"degraded"`
	Profile      string       `json:"profile"`
	Simulated    bool         `json:"simulated"`
	Persistent   bool         `json:"persistent"`
	Dependencies []Dependency `json:"dependencies"`
}
type Health struct {
	mu       sync.Mutex
	snapshot Snapshot
}

func New(c config.Config) *Health {
	storage := Dependency{Name: "storage", Kind: c.Storage, Status: "NOT_IMPLEMENTED", Critical: true, Reason: "SQLite begins on day 16"}
	if c.Storage == "memory_fake" {
		storage.Status = "READY"
		storage.Reason = "simulation only; data is not persisted"
	}
	if c.Storage == "sqlite" {
		storage.Status = "UNAVAILABLE"
		storage.Reason = "awaiting SQLite initialization"
	}
	deps := []Dependency{storage, {Name: "adapters", Kind: "synthetic", Status: "UNAVAILABLE", Critical: true}}
	for _, p := range []struct {
		name    string
		enabled bool
	}{{"ml", c.Components.MLEnabled}, {"mqtt", c.Components.MQTTEnabled}} {
		d := Dependency{Name: p.name, Kind: "external", Status: "DISABLED"}
		if p.enabled {
			d.Status = "NOT_IMPLEMENTED"
			d.Reason = "runtime integration is not implemented"
		}
		deps = append(deps, d)
	}
	if c.Auth.Required {
		deps = append(deps, Dependency{Name: "auth", Kind: "credentials_only", Status: "NOT_IMPLEMENTED", Critical: true, Reason: "sessions begin on day 19"})
	}
	return &Health{snapshot: Snapshot{Live: true, Phase: "STARTING", Profile: c.Profile, Simulated: true, Persistent: false, Dependencies: deps}}
}
func (h *Health) Set(name, status string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for i := range h.snapshot.Dependencies {
		if h.snapshot.Dependencies[i].Name == name {
			h.snapshot.Dependencies[i].Status = status
			if name == "auth" && status == "READY" {
				h.snapshot.Dependencies[i].Kind = "sessions"
				h.snapshot.Dependencies[i].Reason = "session issuer, credential verifier and writable storage initialized"
			}
			if name == "storage" && h.snapshot.Dependencies[i].Kind == "sqlite" {
				h.snapshot.Persistent = status == "READY"
				if status == "READY" {
					h.snapshot.Dependencies[i].Reason = "file-backed SQLite; synthetic adapters remain simulated"
				}
			}
		}
	}
}
func (h *Health) Stopping() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.snapshot.Live = false
	h.snapshot.Phase = "STOPPING"
}
func (h *Health) Snapshot() Snapshot {
	h.mu.Lock()
	defer h.mu.Unlock()
	s := h.snapshot
	s.Dependencies = append([]Dependency(nil), s.Dependencies...)
	s.Ready = s.Live && s.Phase == "RUNNING"
	for _, d := range s.Dependencies {
		if d.Status != "READY" && (d.Critical || d.Status != "DISABLED") {
			if d.Critical {
				s.Ready = false
			} else {
				s.Degraded = true
			}
		}
	}
	return s
}

func (h *Health) SetPhase(phase string) {
	switch phase {
	case "STARTING", "RUNNING", "SAFE_MODE":
	default:
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.snapshot.Live {
		h.snapshot.Phase = phase
	}
}
