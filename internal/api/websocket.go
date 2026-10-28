package api

import (
	"context"
	"github.com/coder/websocket"
	"net/http"
	"time"
)

func (s *Server) websocket(w http.ResponseWriter, r *http.Request) {
	status := 500
	started := time.Now()
	defer func() {
		s.metrics.record("/ws/events", status, time.Since(started))
		s.logger.Info("websocket_closed", "route", "/ws/events", "status", status)
	}()
	w.Header().Set("Cache-Control", "no-store")
	s.wsWait.Add(1)
	defer s.wsWait.Done()
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		status = 405
		jsonResponse(w, 405, map[string]string{"code": "METHOD_NOT_ALLOWED"})
		return
	}
	if r.URL.RawQuery != "" || !s.validOrigin(r, true) {
		status = 403
		jsonResponse(w, 403, map[string]string{"code": "FORBIDDEN"})
		return
	}
	c, err := s.credential(r, false)
	if err != nil {
		status = 401
		jsonResponse(w, 401, map[string]string{"code": "UNAUTHENTICATED"})
		return
	}
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()
	remove, err := s.access.Watch(ctx, c, cancel)
	if err != nil {
		status = 429
		jsonResponse(w, 429, map[string]string{"code": "CONNECTION_LIMIT"})
		return
	}
	defer remove()
	if ctx.Err() != nil {
		status = 401
		jsonResponse(w, 401, map[string]string{"code": "UNAUTHENTICATED"})
		return
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{CompressionMode: websocket.CompressionDisabled})
	if err != nil {
		return
	}
	status = 101
	defer conn.CloseNow()
	conn.SetReadLimit(4096)
	if conn.Write(ctx, websocket.MessageText, []byte(`{"transport":"auth_only","events":false}`)) != nil {
		return
	}
	ctx = conn.CloseRead(ctx)
	ticker := time.NewTicker(s.wsRecheckInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err = s.access.Store.PeekCredential(ctx, c, s.access.Now()); err != nil {
				return
			}
			ping, done := context.WithTimeout(ctx, 3*time.Second)
			err = conn.Ping(ping)
			done()
			if err != nil {
				return
			}
		}
	}
}
