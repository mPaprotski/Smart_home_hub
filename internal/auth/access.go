package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const CookieName = "__Host-shh_session"

func secretHex(size int) (string, error) {
	b := make([]byte, size)
	if _, err := rand.Read(b); err != nil {
		return "", ErrUnavailable
	}
	return hex.EncodeToString(b), nil
}
func digest(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

type requestWindow struct {
	start time.Time
	count int
}
type watcher struct {
	credential storage.Credential
	cancel     context.CancelFunc
}
type Access struct {
	auditFailures atomic.Uint64
	Store         *storage.SQLite
	Credentials   *Service
	Now           func() time.Time
	mu            sync.Mutex
	windows       map[string]requestWindow
	watchers      map[uint64]watcher
	next          uint64
	closed        bool
}

func NewAccess(store *storage.SQLite, service *Service) *Access {
	return &Access{Store: store, Credentials: service, Now: time.Now, windows: make(map[string]requestWindow), watchers: make(map[uint64]watcher)}
}
func (a *Access) Issue(ctx context.Context, user string) (raw, csrf string, err error) {
	raw, err = secretHex(32)
	if err != nil {
		return "", "", err
	}
	csrf, err = secretHex(32)
	if err != nil {
		return "", "", err
	}
	if err = a.Store.IssueSession(ctx, user, digest(raw), csrf, a.Now()); err != nil {
		return "", "", err
	}
	a.Recheck(ctx)
	return raw, csrf, nil
}
func (a *Access) Verify(ctx context.Context, session, bearer, peer string) (result storage.Credential, err error) {
	defer func() {
		if bearer != "" && err != nil {
			if e := a.Store.AuditDenied(ctx, a.Now(), "token_denied"); e != nil {
				a.auditFailures.Add(1)
			}
		}
	}()
	if session != "" && bearer != "" {
		return storage.Credential{}, storage.ErrInvalid
	}
	var c storage.Credential
	if bearer != "" {
		parts := strings.Split(bearer, ".")
		if len(parts) != 2 || len(parts[0]) != 32 || len(parts[1]) != 64 {
			return c, ErrInvalid
		}
		if _, err := hex.DecodeString(parts[0] + parts[1]); err != nil {
			return c, ErrInvalid
		}
		c = storage.Credential{Kind: "token", ID: parts[0], Proof: digest(parts[1])}
		if !a.allow("token:"+c.ID, 60) || !a.allow("peer:"+peer, 120) {
			return c, ErrRate
		}
	} else {
		if len(session) != 64 {
			return c, ErrInvalid
		}
		if _, err := hex.DecodeString(session); err != nil {
			return c, ErrInvalid
		}
		c = storage.Credential{Kind: "session", ID: digest(session)}
	}
	actual, err := a.Store.CheckCredential(ctx, c, a.Now())
	if err != nil {
		return storage.Credential{}, err
	}
	return actual, nil
}
func (a *Access) allow(key string, limit int) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.Now()
	for key, w := range a.windows {
		if !now.Before(w.start.Add(time.Minute)) {
			delete(a.windows, key)
		}
	}
	w, exists := a.windows[key]
	if !exists {
		if len(a.windows) >= 4096 {
			return false
		}
		w.start = now
	}
	if w.count >= limit {
		return false
	}
	w.count++
	a.windows[key] = w
	return true
}
func ValidCSRF(c storage.Credential, header string) bool {
	return c.Kind == "session" && len(header) == 64 && subtle.ConstantTimeCompare([]byte(c.CSRF), []byte(header)) == 1
}
func (a *Access) Revoke(ctx context.Context, c storage.Credential, all bool) error {
	if err := a.Store.RevokeSessions(ctx, c, all, a.Now()); err != nil {
		return err
	}
	a.Recheck(ctx)
	return nil
}
func (a *Access) CreateToken(ctx context.Context, c storage.Credential, name string, scopes, bindings []string) (storage.TokenMetadata, string, error) {
	var meta storage.TokenMetadata
	if len(name) < 1 || len(name) > 64 || len(scopes) == 0 || len(scopes) > 8 || len(bindings) == 0 || len(bindings) > 16 {
		return meta, "", storage.ErrInvalid
	}
	for _, r := range name {
		if r < 32 || r > 126 {
			return meta, "", storage.ErrInvalid
		}
	}
	seen := map[string]bool{}
	for _, v := range append(append([]string{}, scopes...), bindings...) {
		if len(v) > 128 || seen[v] {
			return meta, "", storage.ErrInvalid
		}
		seen[v] = true
	}
	id, err := secretHex(16)
	if err != nil {
		return meta, "", err
	}
	raw, err := secretHex(32)
	if err != nil {
		return meta, "", err
	}
	now := a.Now()
	meta = storage.TokenMetadata{ID: id, Name: name, Scopes: scopes, Bindings: bindings, CreatedAt: now.Unix(), ExpiresAt: now.Add(90 * 24 * time.Hour).Unix()}
	if err = a.Store.CreateToken(ctx, c, meta, digest(raw), now); err != nil {
		return storage.TokenMetadata{}, "", err
	}
	return meta, id + "." + raw, nil
}
func (a *Access) RevokeToken(ctx context.Context, c storage.Credential, id string) error {
	if len(id) != 32 {
		return storage.ErrInvalid
	}
	if err := a.Store.RevokeToken(ctx, c, id, a.Now()); err != nil {
		return err
	}
	a.Recheck(ctx)
	return nil
}

// Watch registration and a second credential check close the verify/register race.
func (a *Access) Watch(ctx context.Context, c storage.Credential, cancel context.CancelFunc) (func(), error) {
	a.mu.Lock()
	if a.closed || len(a.watchers) >= 16 {
		a.mu.Unlock()
		return nil, ErrRate
	}
	count := 0
	for _, w := range a.watchers {
		if w.credential.Kind == c.Kind && w.credential.ID == c.ID {
			count++
		}
	}
	if count >= 2 {
		a.mu.Unlock()
		return nil, ErrRate
	}
	a.next++
	id := a.next
	a.watchers[id] = watcher{c, cancel}
	a.mu.Unlock()
	remove := func() { a.mu.Lock(); delete(a.watchers, id); a.mu.Unlock() }
	if _, err := a.Store.PeekCredential(ctx, c, a.Now()); err != nil {
		remove()
		cancel()
		return nil, err
	}
	return remove, nil
}
func (a *Access) Recheck(ctx context.Context) {
	a.mu.Lock()
	snapshot := make(map[uint64]watcher, len(a.watchers))
	for id, w := range a.watchers {
		snapshot[id] = w
	}
	a.mu.Unlock()
	for id, w := range snapshot {
		if _, err := a.Store.PeekCredential(ctx, w.credential, a.Now()); err != nil {
			w.cancel()
			a.mu.Lock()
			delete(a.watchers, id)
			a.mu.Unlock()
		}
	}
}
func (a *Access) Close() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.closed = true
	for id, w := range a.watchers {
		w.cancel()
		delete(a.watchers, id)
	}
}
func (a *Access) Connections() int { a.mu.Lock(); defer a.mu.Unlock(); return len(a.watchers) }
func PublicAuthError(err error) (int, string) {
	if errors.Is(err, ErrRate) {
		return 429, "RATE_LIMITED"
	}
	if errors.Is(err, storage.ErrInvalid) {
		return 400, "INVALID_REQUEST"
	}
	if errors.Is(err, ErrInvalid) || err == storage.Error("UNAUTHENTICATED") {
		return 401, "UNAUTHENTICATED"
	}
	if err == storage.Error("FORBIDDEN") {
		return 403, "FORBIDDEN"
	}
	if err == storage.ErrConflict {
		return 409, "REVISION_CONFLICT"
	}
	if err == storage.ErrNotFound {
		return 404, "NOT_FOUND"
	}
	if err == storage.Error("TOKEN_LIMIT") {
		return 429, "TOKEN_LIMIT"
	}
	return 503, "AUTH_UNAVAILABLE"
}

func (a *Access) AuditFailures() uint64 { return a.auditFailures.Load() }
