package api

import (
	"context"
	"encoding/json"
	"github.com/mPaprotski/Smart_home_hub/internal/config"
	"github.com/mPaprotski/Smart_home_hub/internal/events"
	"github.com/mPaprotski/Smart_home_hub/internal/system"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func fixture(storage string) (*Server, *system.Health) {
	c := config.Defaults()
	c.Server.Port = 0
	c.Storage = storage
	h := system.New(c)
	h.SetPhase("RUNNING")
	return New(c, h, events.New(8, 4096), slog.New(slog.NewJSONHandler(io.Discard, nil)), make(chan error, 1)), h
}
func TestRoutesReadinessAndMethods(t *testing.T) {
	s, h := fixture("not_implemented")
	h.Set("adapters", "READY")
	for _, c := range []struct {
		method, path string
		status       int
	}{{"GET", "/health", 200}, {"GET", "/ready", 503}, {"GET", "/api/v1/version", 200}, {"POST", "/api/v1/version", 405}, {"POST", "/api/v1/commands", 404}} {
		r := httptest.NewRequest(c.method, c.path, nil)
		r.Header.Set("X-Request-ID", "invalid token!")
		w := httptest.NewRecorder()
		s.Handler().ServeHTTP(w, r)
		if w.Code != c.status || !requestID.MatchString(w.Header().Get("X-Request-ID")) {
			t.Fatal(c, w.Code)
		}
	}
	s, h = fixture("memory_fake")
	h.Set("adapters", "READY")
	h.Set("ml", "UNAVAILABLE")
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/ready", nil))
	var snap system.Snapshot
	if err := json.Unmarshal(w.Body.Bytes(), &snap); err != nil || w.Code != 200 || !snap.Degraded || snap.Persistent || !snap.Simulated {
		t.Fatal(w.Body.String(), err)
	}
	h.Stopping()
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/ready", nil))
	if w.Code != 503 {
		t.Fatal(w.Code)
	}
}
func TestBoundedMetricsAndRequestIDs(t *testing.T) {
	s, _ := fixture("memory_fake")
	for _, path := range []string{"/arbitrary/one?secret=never-log", "/arbitrary/two", "/health"} {
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", path, nil)
		r.Header.Set("X-Request-ID", "my_request_1")
		s.Handler().ServeHTTP(w, r)
		if w.Header().Get("X-Request-ID") != "my_request_1" {
			t.Fatal("id lost")
		}
	}
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/metrics", nil))
	out := w.Body.String()
	if strings.Contains(out, "arbitrary") || strings.Contains(out, "secret") || !strings.Contains(out, `hub_http_requests_total{route="unmatched"} 2`) || !strings.Contains(out, "hub_http_request_duration_ms_bucket") {
		t.Fatal(out)
	}
}
func TestActualServerShutdown(t *testing.T) {
	s, h := fixture("memory_fake")
	h.Set("adapters", "READY")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := s.Start(ctx); err != nil {
		t.Fatal(err)
	}
	client := http.Client{Timeout: time.Second}
	resp, err := client.Get("http://" + s.Address() + "/ready")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
	stop, done := context.WithTimeout(context.Background(), time.Second)
	defer done()
	if err := s.Stop(stop); err != nil {
		t.Fatal(err)
	}
	if _, err := client.Get("http://" + s.Address() + "/health"); err == nil {
		t.Fatal("port still accepting")
	}
}

func TestPanicRecoveryAndListenFailure(t *testing.T) {
	s, _ := fixture("memory_fake")
	s.health = nil
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/health", nil))
	if w.Code != 500 || !strings.Contains(w.Body.String(), "INTERNAL") || strings.Contains(w.Body.String(), "panic") {
		t.Fatal(w.Code, w.Body.String())
	}
	s, _ = fixture("memory_fake")
	if err := s.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer s.Stop(context.Background())
	other, _ := fixture("memory_fake")
	other.cfg.Server.Host = "127.0.0.1"
	addr := s.listener.Addr().(*net.TCPAddr)
	other.cfg.Server.Port = addr.Port
	if other.Start(context.Background()) == nil {
		t.Fatal("duplicate listener accepted")
	}
}
