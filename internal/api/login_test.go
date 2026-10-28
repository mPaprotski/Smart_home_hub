package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/mPaprotski/Smart_home_hub/internal/auth"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestActualCredentialLoginStageAndRedaction(t *testing.T) {
	db, err := storage.Open(context.Background(), storage.Options{Path: filepath.Join(t.TempDir(), "data", "hub.db"), BusyTimeout: time.Second, StartupTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	service, err := auth.New(context.Background(), db, auth.NewHasher())
	if err != nil {
		t.Fatal(err)
	}
	secret := "Synthetic password sentinel 42"
	if _, err = service.Bootstrap(context.Background(), "admin", []byte(secret)); err != nil {
		t.Fatal(err)
	}
	server, _ := fixture("memory_fake")
	server.SetAuthenticator(service)
	for _, c := range []struct {
		body   string
		code   int
		result string
	}{{`{"username":"admin","password":"` + secret + `"}`, 503, "AUTH_SESSION_NOT_IMPLEMENTED"}, {`{"username":"admin","password":"wrong"}`, 401, "INVALID_CREDENTIALS"}, {`{"username":"missing","password":"wrong"}`, 401, "INVALID_CREDENTIALS"}, {`{"username":"admin","password":"wrong","principal":"admin"}`, 400, "INVALID_REQUEST"}, {`{"username":"admin","password":"wrong"} {}`, 400, "INVALID_REQUEST"}, {`{"username":"admin","password":"` + strings.Repeat("x", 5000) + `"}`, 400, "INVALID_REQUEST"}} {
		r := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(c.body))
		r.Header.Set("Content-Type", "application/json")
		r.RemoteAddr = "127.0.0.1:1234"
		w := httptest.NewRecorder()
		server.Handler().ServeHTTP(w, r)
		var body map[string]string
		if err = json.Unmarshal(w.Body.Bytes(), &body); err != nil || w.Code != c.code || body["code"] != c.result || w.Header().Get("Set-Cookie") != "" || strings.Contains(w.Body.String(), secret) {
			t.Fatal(w.Code, w.Body.String(), err)
		}
	}
}

type spyAuth struct {
	peer string
	err  error
}

func (a *spyAuth) Authenticate(_ context.Context, _ string, _ []byte, peer string) (auth.Principal, error) {
	a.peer = peer
	return auth.Principal{}, a.err
}
func TestLoginPeerMethodAndLimits(t *testing.T) {
	s, _ := fixture("memory_fake")
	spy := &spyAuth{err: auth.ErrRate}
	s.SetAuthenticator(spy)
	r := httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"wrong"}`))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-Forwarded-For", "1.2.3.4")
	r.RemoteAddr = "127.0.0.1:1234"
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 429 || spy.peer != "127.0.0.1" || w.Header().Get("Retry-After") == "" {
		t.Fatal(w.Code, spy.peer)
	}
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, httptest.NewRequest("GET", "/api/v1/auth/login", nil))
	if w.Code != 405 || w.Header().Get("Allow") != "POST" {
		t.Fatal(w.Code)
	}
	spy.err = errors.New("secret database path")
	r = httptest.NewRequest("POST", "/api/v1/auth/login", strings.NewReader(`{"username":"admin","password":"wrong"}`))
	r.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 503 || strings.Contains(w.Body.String(), "secret") {
		t.Fatal(w.Body.String())
	}
}
