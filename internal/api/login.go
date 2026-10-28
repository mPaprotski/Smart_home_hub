package api

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/mPaprotski/Smart_home_hub/internal/auth"
	"io"
	"mime"
	"net"
	"net/http"
	"time"
)

type Authenticator interface {
	Authenticate(context.Context, string, []byte, string) (auth.Principal, error)
}

func (s *Server) SetAuthenticator(a Authenticator) { s.auth = a } // composition time, before Start
func (s *Server) login(w http.ResponseWriter, r *http.Request, id string) int {
	respond := func(status int, code string) int {
		jsonResponse(w, status, map[string]string{"code": code, "request_id": id})
		return status
	}
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		return respond(405, "METHOD_NOT_ALLOWED")
	}
	if s.access != nil && (!s.validOrigin(r, true) || r.Header.Get("X-Hub-Request") != "1") {
		return respond(403, "FORBIDDEN")
	}
	media, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || media != "application/json" {
		return respond(400, "INVALID_REQUEST")
	}
	var body struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&body) != nil {
		return respond(400, "INVALID_REQUEST")
	}
	var extra any
	if err = decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return respond(400, "INVALID_REQUEST")
	}
	if len(body.Username) < 1 || len(body.Username) > 128 || len(body.Password) < 1 || len(body.Password) > 1024 {
		return respond(400, "INVALID_REQUEST")
	}
	if s.auth == nil {
		return respond(503, "AUTH_UNAVAILABLE")
	}
	peer, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return respond(400, "INVALID_REQUEST")
	}
	principal, err := s.auth.Authenticate(r.Context(), body.Username, []byte(body.Password), peer)
	switch {
	case errors.Is(err, auth.ErrInvalid):
		if s.access != nil {
			if auditErr := s.access.Store.AuditFailedLogin(r.Context(), s.access.Now()); auditErr != nil {
				s.auditFailures.Add(1)
			}
		}
		return respond(401, "INVALID_CREDENTIALS")
	case errors.Is(err, auth.ErrRate):
		w.Header().Set("Retry-After", "60")
		return respond(429, "RATE_LIMITED")
	case err != nil:
		return respond(503, "AUTH_UNAVAILABLE")
	default:
		if s.access == nil {
			return respond(503, "AUTH_SESSION_NOT_IMPLEMENTED")
		}
		raw, csrf, err := s.access.Issue(r.Context(), principal.UserID)
		if err != nil {
			return respond(503, "AUTH_UNAVAILABLE")
		}
		s.notifyAudit()
		http.SetCookie(w, &http.Cookie{Name: auth.CookieName, Value: raw, Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 43200, Expires: s.access.Now().Add(12 * time.Hour)})
		jsonResponse(w, 200, map[string]string{"user_id": principal.UserID, "csrf_token": csrf})
		return 200
	}
}
