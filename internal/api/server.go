// Package api exposes read-only development diagnostics. Authentication and commands
// are deliberately unavailable until their implementation stages.
package api

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mPaprotski/Smart_home_hub/internal/auth"
	"github.com/mPaprotski/Smart_home_hub/internal/buildinfo"
	"github.com/mPaprotski/Smart_home_hub/internal/commands"
	"github.com/mPaprotski/Smart_home_hub/internal/config"
	"github.com/mPaprotski/Smart_home_hub/internal/events"
	"github.com/mPaprotski/Smart_home_hub/internal/system"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"time"
)

type routeMetric struct {
	Count   uint64
	Errors  uint64
	SumMS   float64
	Buckets [6]uint64
}

var bounds = [6]float64{1, 5, 10, 50, 100, 1000}

type Metrics struct {
	mu     sync.Mutex
	routes map[string]*routeMetric
}

func (m *Metrics) record(route string, status int, elapsed time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.routes == nil {
		m.routes = make(map[string]*routeMetric)
	}
	r := m.routes[route]
	if r == nil {
		r = &routeMetric{}
		m.routes[route] = r
	}
	r.Count++
	if status >= 500 {
		r.Errors++
	}
	ms := float64(elapsed) / float64(time.Millisecond)
	r.SumMS += ms
	for i, b := range bounds {
		if ms <= b {
			r.Buckets[i]++
		}
	}
}
func (m *Metrics) write(w http.ResponseWriter, bus *events.Bus) {
	m.mu.Lock()
	defer m.mu.Unlock()
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintln(w, "# TYPE hub_http_requests_total counter\n# TYPE hub_http_errors_total counter\n# TYPE hub_http_request_duration_ms histogram")
	for _, route := range []string{"/health", "/ready", "/metrics", "/api/v1/version", "/api/v1/auth/login", "/api/v1/auth/me", "/api/v1/auth/csrf", "/api/v1/auth/logout", "/api/v1/auth/revoke-all", "/api/v1/auth/tokens", "/api/v1/auth/tokens/revoke", "/api/v1/auth/audit", "/api/v1/users/me/settings/timezone", "/ws/events", "unmatched"} {
		r := m.routes[route]
		if r == nil {
			continue
		}
		fmt.Fprintf(w, "hub_http_requests_total{route=%q} %d\nhub_http_errors_total{route=%q} %d\n", route, r.Count, route, r.Errors)
		for i, b := range bounds {
			fmt.Fprintf(w, "hub_http_request_duration_ms_bucket{route=%q,le=\"%g\"} %d\n", route, b, r.Buckets[i])
		}
		fmt.Fprintf(w, "hub_http_request_duration_ms_bucket{route=%q,le=\"+Inf\"} %d\nhub_http_request_duration_ms_sum{route=%q} %g\nhub_http_request_duration_ms_count{route=%q} %d\n", route, r.Count, route, r.SumMS, route, r.Count)
	}
	s := bus.Stats()
	fmt.Fprintf(w, "# TYPE hub_events_published_total counter\nhub_events_published_total %d\nhub_events_dropped_total %d\nhub_events_overflow_total %d\nhub_events_rejected_total %d\nhub_event_subscribers %d\n", s.Published, s.Dropped, s.Overflows, s.Rejected, s.Subscribers)
}

type Server struct {
	auditFailures        atomic.Uint64
	notificationSequence atomic.Uint64
	notificationFailures atomic.Uint64
	notificationEpoch    string
	cfg                  config.Config
	health               *system.Health
	bus                  *events.Bus
	logger               *slog.Logger
	metrics              Metrics
	mu                   sync.Mutex
	server               *http.Server
	listener             net.Listener
	done                 chan struct{}
	auth                 Authenticator
	access               *auth.Access
	dispatcher           *commands.Dispatcher
	wsWait               sync.WaitGroup
	wsRecheckInterval    time.Duration
	fatal                chan<- error
}

func New(c config.Config, h *system.Health, b *events.Bus, logger *slog.Logger, fatal chan<- error) *Server {
	return &Server{cfg: c, health: h, bus: b, logger: logger, fatal: fatal, notificationEpoch: fmt.Sprintf("api-%d", time.Now().UnixNano()), wsRecheckInterval: 15 * time.Second}
}
func (s *Server) Address() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.listener == nil {
		return ""
	}
	return s.listener.Addr().String()
}
func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.server != nil {
		return errors.New("HTTP server reused")
	}
	listener, err := net.Listen("tcp", s.cfg.Address())
	if err != nil {
		return errors.New("HTTP listen failed")
	}
	if s.cfg.Server.TLSCert != "" {
		pair, err := tls.LoadX509KeyPair(s.cfg.Server.TLSCert, s.cfg.Server.TLSKey)
		if err != nil {
			_ = listener.Close()
			return errors.New("TLS certificate rejected")
		}
		listener = tls.NewListener(listener, &tls.Config{Certificates: []tls.Certificate{pair}, MinVersion: tls.VersionTLS12})
	}
	readHeader, _ := time.ParseDuration(s.cfg.Server.ReadHeaderTimeout)
	s.listener = listener
	s.done = make(chan struct{})
	s.server = &http.Server{Handler: s.Handler(), ReadHeaderTimeout: readHeader, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 30 * time.Second, MaxHeaderBytes: 16 << 10, BaseContext: func(net.Listener) context.Context { return ctx }}
	go func() {
		defer close(s.done)
		err := s.server.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			select {
			case s.fatal <- errors.New("HTTP serve failed"):
			case <-ctx.Done():
			}
		}
	}()
	s.logger.Info("http_started", "address", listener.Addr().String(), "profile", s.cfg.Profile, "simulated", true)
	return nil
}
func (s *Server) Stop(ctx context.Context) error {
	s.health.Stopping()
	if s.access != nil {
		s.access.Close()
	}
	s.mu.Lock()
	server, done := s.server, s.done
	s.mu.Unlock()
	if server == nil {
		return nil
	}
	err := server.Shutdown(ctx)
	if err != nil {
		_ = server.Close()
	}
	select {
	case <-done:
		finished := make(chan struct{})
		go func() { s.wsWait.Wait(); close(finished) }()
		select {
		case <-finished:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	case <-ctx.Done():
		return ctx.Err()
	}
}

var requestID = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func jsonResponse(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func (s *Server) Handler() http.Handler {
	return http.HandlerFunc(func(dst http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ws/events" && s.access != nil {
			s.websocket(dst, r)
			return
		}
		buffer := &responseBuffer{header: make(http.Header)}
		var w http.ResponseWriter = buffer
		started := time.Now()
		id := r.Header.Get("X-Request-ID")
		if !requestID.MatchString(id) {
			var b [16]byte
			if _, err := rand.Read(b[:]); err != nil {
				jsonResponse(w, 500, map[string]string{"code": "INTERNAL"})
				buffer.flush(dst)
				return
			}
			id = hex.EncodeToString(b[:])
		}
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		route := r.URL.Path
		switch route {
		case "/health", "/ready", "/metrics", "/api/v1/version", "/api/v1/auth/login":
		case "/api/v1/auth/me", "/api/v1/auth/csrf", "/api/v1/auth/logout", "/api/v1/auth/revoke-all", "/api/v1/auth/tokens", "/api/v1/auth/tokens/revoke", "/api/v1/auth/audit", "/api/v1/users/me/settings/timezone":
			if s.access == nil {
				route = "unmatched"
			}
		default:
			route = "unmatched"
		}
		status := 200
		defer func() {
			if recover() != nil {
				status = 500
				buffer.body.Reset()
				buffer.status = 0
				jsonResponse(w, status, map[string]string{"code": "INTERNAL", "request_id": id})
			}
			buffer.flush(dst)
			s.metrics.record(route, status, time.Since(started))
			s.logger.Info("http_request", "route", route, "status", status, "duration_ms", float64(time.Since(started))/float64(time.Millisecond), "request_id", id)
		}()
		if route == "unmatched" {
			status = 404
			jsonResponse(w, status, map[string]string{"code": "NOT_FOUND", "request_id": id})
			return
		}
		if route == "/api/v1/auth/login" {
			status = s.login(w, r, id)
			return
		}
		if s.access != nil && route != "/api/v1/version" && route != "/health" && route != "/ready" && route != "/metrics" {
			status = s.accessRoute(w, r, id)
			return
		}
		if r.Method != http.MethodGet {
			status = 405
			w.Header().Set("Allow", "GET")
			jsonResponse(w, status, map[string]string{"code": "METHOD_NOT_ALLOWED", "request_id": id})
			return
		}
		switch route {
		case "/health":
			snap := s.health.Snapshot()
			if !snap.Live {
				status = 503
			}
			jsonResponse(w, status, map[string]any{"live": snap.Live, "profile": snap.Profile, "simulated": snap.Simulated})
		case "/ready":
			snap := s.health.Snapshot()
			if !snap.Ready {
				status = 503
			}
			jsonResponse(w, status, snap)
		case "/metrics":
			s.metrics.write(w, s.bus)
			failures := s.auditFailures.Load()
			if s.access != nil {
				failures += s.access.AuditFailures()
			}
			fmt.Fprintf(w, "hub_audit_write_failures_total %d\nhub_audit_notifications_failed_total %d\n", failures, s.notificationFailures.Load())
		case "/api/v1/version":
			jsonResponse(w, status, map[string]any{"api_version": "v1", "version": buildinfo.Version, "commit": buildinfo.Commit, "profile": s.cfg.Profile, "adapter_mode": s.cfg.AdapterMode, "storage_kind": s.cfg.Storage, "simulated": true, "persistent": s.health.Snapshot().Persistent, "auth": s.authStage()})
		}
	})
}

type responseBuffer struct {
	header http.Header
	body   bytes.Buffer
	status int
}

func (b *responseBuffer) Header() http.Header { return b.header }
func (b *responseBuffer) WriteHeader(status int) {
	if b.status == 0 {
		b.status = status
	}
}
func (b *responseBuffer) Write(p []byte) (int, error) {
	if b.status == 0 {
		b.status = 200
	}
	if b.body.Len()+len(p) > 65536 {
		panic("response size limit")
	}
	return b.body.Write(p)
}
func (b *responseBuffer) flush(dst http.ResponseWriter) {
	for k, v := range b.header {
		dst.Header()[k] = v
	}
	if b.status == 0 {
		b.status = 200
	}
	dst.WriteHeader(b.status)
	_, _ = dst.Write(b.body.Bytes())
}

func (s *Server) authStage() string {
	if s.access != nil {
		return "SESSIONS_AND_TOKENS"
	}
	if s.auth != nil {
		return "CREDENTIALS_ONLY"
	}
	return "NOT_IMPLEMENTED"
}
