package api

import (
	"encoding/json"
	"errors"
	"github.com/mPaprotski/Smart_home_hub/internal/auth"
	"github.com/mPaprotski/Smart_home_hub/internal/commands"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func (s *Server) SetAccess(a *auth.Access) { s.access = a; s.dispatcher = commands.New(a.Store) } // before Start
func (s *Server) secure(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(peer)
	for _, cidr := range s.cfg.Auth.TrustedProxies {
		_, network, e := net.ParseCIDR(cidr)
		if e == nil && network.Contains(ip) {
			return len(r.Header.Values("X-Forwarded-Proto")) == 1 && r.Header.Get("X-Forwarded-Proto") == "https"
		}
	}
	return false
}
func (s *Server) validOrigin(r *http.Request, required bool) bool {
	if !s.secure(r) {
		return false
	}
	u, err := url.Parse(s.cfg.Auth.PublicOrigin)
	if err != nil || u.Scheme != "https" || u.Host != r.Host {
		return false
	}
	values := r.Header.Values("Origin")
	if len(values) == 0 {
		return !required
	}
	return len(values) == 1 && values[0] == s.cfg.Auth.PublicOrigin
}
func readJSON(w http.ResponseWriter, r *http.Request, dest any) error {
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return storage.ErrInvalid
	}
	d := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	d.DisallowUnknownFields()
	if d.Decode(dest) != nil {
		return storage.ErrInvalid
	}
	var extra any
	if !errors.Is(d.Decode(&extra), io.EOF) {
		return storage.ErrInvalid
	}
	return nil
}
func rawCredential(r *http.Request) (session, bearer string, err error) {
	count := 0
	for _, cookie := range r.Cookies() {
		if cookie.Name == auth.CookieName {
			count++
			session = cookie.Value
		}
	}
	headers := r.Header.Values("Authorization")
	if count > 1 || len(headers) > 1 {
		return "", "", storage.ErrInvalid
	}
	if len(headers) == 1 {
		if !strings.HasPrefix(headers[0], "Bearer ") {
			return "", "", storage.ErrInvalid
		}
		bearer = strings.TrimPrefix(headers[0], "Bearer ")
		if bearer == "" {
			return "", "", storage.ErrInvalid
		}
	}
	if session != "" && bearer != "" {
		return "", "", storage.ErrInvalid
	}
	return session, bearer, nil
}
func (s *Server) credential(r *http.Request, mutating bool) (storage.Credential, error) {
	if s.access == nil {
		return storage.Credential{}, auth.ErrUnavailable
	}
	session, bearer, err := rawCredential(r)
	if err != nil {
		return storage.Credential{}, err
	}
	if !s.secure(r) {
		return storage.Credential{}, auth.ErrUnavailable
	}
	if bearer == "" && !s.validOrigin(r, mutating) {
		return storage.Credential{}, storage.Error("FORBIDDEN")
	}
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return storage.Credential{}, storage.ErrInvalid
	}
	c, err := s.access.Verify(r.Context(), session, bearer, peer)
	if err != nil {
		return c, err
	}
	if mutating && c.Kind == "session" && !auth.ValidCSRF(c, r.Header.Get("X-CSRF-Token")) {
		return storage.Credential{}, storage.Error("FORBIDDEN")
	}
	return c, nil
}
func (s *Server) accessRoute(w http.ResponseWriter, r *http.Request, id string) int {
	respond := func(status int, value any) int {
		if status == 204 {
			w.WriteHeader(status)
		} else {
			jsonResponse(w, status, value)
		}
		return status
	}
	failure := func(err error) int {
		if s.access != nil && r.URL.Path == "/api/v1/users/me/settings/timezone" {
			if e := s.access.Store.AuditDenied(r.Context(), s.access.Now(), "setting_update"); e != nil {
				s.auditFailures.Add(1)
			}
		}
		code, label := auth.PublicAuthError(err)
		if code == 429 {
			w.Header().Set("Retry-After", "60")
		}
		return respond(code, map[string]string{"code": label, "request_id": id})
	}
	method := "GET"
	path := r.URL.Path
	switch path {
	case "/api/v1/auth/logout", "/api/v1/auth/revoke-all", "/api/v1/auth/tokens/revoke":
		method = "POST"
	case "/api/v1/users/me/settings/timezone":
		method = "PUT"
	case "/api/v1/auth/tokens":
		if r.Method == "POST" {
			method = "POST"
		}
	}
	if r.Method != method {
		w.Header().Set("Allow", method)
		return respond(405, map[string]string{"code": "METHOD_NOT_ALLOWED", "request_id": id})
	}
	c, err := s.credential(r, method != "GET")
	if err != nil {
		return failure(err)
	}
	switch path {
	case "/api/v1/auth/me":
		return respond(200, map[string]any{"user_id": c.UserID, "credential_kind": c.Kind})
	case "/api/v1/auth/csrf":
		if c.Kind != "session" {
			return failure(storage.Error("FORBIDDEN"))
		}
		return respond(200, map[string]string{"csrf_token": c.CSRF})
	case "/api/v1/auth/logout", "/api/v1/auth/revoke-all":
		if err = s.access.Revoke(r.Context(), c, path == "/api/v1/auth/revoke-all"); err != nil {
			return failure(err)
		}
		s.notifyAudit()
		http.SetCookie(w, &http.Cookie{Name: auth.CookieName, Value: "", Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: -1, Expires: time.Unix(1, 0)})
		return respond(204, nil)
	case "/api/v1/auth/tokens":
		if method == "GET" {
			tokens, err := s.access.Store.ListTokens(r.Context(), c, s.access.Now())
			if err != nil {
				return failure(err)
			}
			return respond(200, tokens)
		}
		var body struct {
			Name     string   `json:"name"`
			Scopes   []string `json:"scopes"`
			Bindings []string `json:"bindings"`
		}
		if readJSON(w, r, &body) != nil {
			return failure(storage.ErrInvalid)
		}
		meta, raw, err := s.access.CreateToken(r.Context(), c, body.Name, body.Scopes, body.Bindings)
		if err != nil {
			return failure(err)
		}
		s.notifyAudit()
		return respond(201, map[string]any{"metadata": meta, "token": raw})
	case "/api/v1/auth/tokens/revoke":
		var body struct {
			ID string `json:"id"`
		}
		if readJSON(w, r, &body) != nil {
			return failure(storage.ErrInvalid)
		}
		if err = s.access.RevokeToken(r.Context(), c, body.ID); err != nil {
			return failure(err)
		}
		s.notifyAudit()
		return respond(204, nil)
	case "/api/v1/auth/audit":
		after := int64(0)
		if raw := r.URL.Query().Get("after"); raw != "" {
			after, err = strconv.ParseInt(raw, 10, 64)
			if err != nil || after < 0 {
				return failure(storage.ErrInvalid)
			}
		}
		rows, err := s.access.Store.ListAudit(r.Context(), c, after, s.access.Now())
		if err != nil {
			return failure(err)
		}
		return respond(200, rows)
	case "/api/v1/users/me/settings/timezone":
		var body struct {
			Value            string `json:"value"`
			ExpectedRevision *int64 `json:"expected_revision"`
			CommandID        string `json:"command_id"`
		}
		if readJSON(w, r, &body) != nil {
			return failure(storage.ErrInvalid)
		}
		if body.ExpectedRevision == nil || body.Value == "" || body.CommandID == "" {
			return failure(storage.ErrInvalid)
		}
		p := commands.Principal{ID: c.UserID, UserID: c.UserID, Kind: "user", Credential: &c}
		source := "WEB_UI"
		if c.Kind == "token" {
			p.ID = c.ID
			p.Kind = "service"
			p.Scopes = c.Scopes
			p.Bindings = c.Bindings
			source = "EXTERNAL_API"
		}
		out, err := s.dispatcher.Execute(r.Context(), p, source, commands.Command{ID: body.CommandID, Version: 1, Type: "SetUserSetting", Target: c.UserID, Key: "timezone", Value: body.Value, ExpectedRevision: *body.ExpectedRevision, ExpiresAt: s.access.Now().Add(5 * time.Second), CorrelationID: id})
		if err != nil {
			if err == commands.ErrForbidden {
				err = storage.Error("FORBIDDEN")
			}
			if err == commands.ErrInvalid {
				err = storage.ErrInvalid
			}
			if err == commands.ErrConflict {
				return respond(409, map[string]string{"code": "IDEMPOTENCY_CONFLICT", "request_id": id})
			}
			if err == storage.ErrConflict {
				return respond(412, map[string]string{"code": "REVISION_CONFLICT", "request_id": id})
			}
			return failure(err)
		}
		s.notifyAudit()
		return respond(200, out)
	}
	return respond(404, map[string]string{"code": "NOT_FOUND", "request_id": id})
}
