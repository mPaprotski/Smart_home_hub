// Package auth verifies credentials only; sessions and transport gateways are later stages.
package auth

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"sync"
	"time"
)

type Principal struct{ UserID string }
type ipWindow struct {
	start time.Time
	count int
}
type Service struct {
	store  *storage.SQLite
	hasher *Hasher
	dummy  string
	key    []byte
	clock  func() time.Time
	mu     sync.Mutex
	ips    map[string]ipWindow
	active chan struct{}
}

func New(ctx context.Context, store *storage.SQLite, hasher *Hasher) (*Service, error) {
	key, err := store.EnsureLimiterKey(ctx)
	if err != nil {
		return nil, ErrUnavailable
	}
	dummy, err := hasher.Hash(ctx, []byte("synthetic-dummy-password-only"))
	if err != nil {
		return nil, err
	}
	return &Service{store: store, hasher: hasher, dummy: dummy, key: key, clock: time.Now, ips: make(map[string]ipWindow), active: make(chan struct{}, 1)}, nil
}
func (s *Service) allowIP(ip string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, w := range s.ips {
		if !now.Before(w.start.Add(time.Minute)) {
			delete(s.ips, k)
		}
	}
	w, exists := s.ips[ip]
	if !exists && len(s.ips) >= 4096 {
		return false
	}
	if w.count >= 20 {
		return false
	}
	if !exists {
		w.start = now
	}
	w.count++
	s.ips[ip] = w
	return true
}
func (s *Service) Authenticate(ctx context.Context, username string, password []byte, peerIP string) (Principal, error) {
	now := s.clock().UTC()
	if len(username) > 128 || len(peerIP) > 64 || len(password) == 0 || len(password) > 1024 {
		return Principal{}, ErrInvalid
	}
	if !s.allowIP(peerIP, now) {
		return Principal{}, ErrRate
	}
	// Covers repository work as well as Argon2; no queue of password-bearing jobs.
	select {
	case s.active <- struct{}{}:
		defer func() { <-s.active }()
	default:
		return Principal{}, ErrRate
	}
	name, nameErr := storage.NormalizeUsername(username)
	if nameErr != nil {
		name = "invalid_username"
	}
	mac := hmac.New(sha256.New, s.key)
	_, _ = mac.Write([]byte(name))
	account := hex.EncodeToString(mac.Sum(nil))
	allowed, err := s.store.Attempt(ctx, account, now, false)
	if err != nil {
		return Principal{}, ErrUnavailable
	}
	if !allowed {
		return Principal{}, ErrRate
	}
	user, findErr := s.store.FindUser(ctx, name)
	if findErr != nil && !errors.Is(findErr, storage.ErrNotFound) {
		return Principal{}, ErrUnavailable
	}
	encoded := s.dummy
	if nameErr == nil && findErr == nil && user.Enabled {
		if _, err := parseHash(user.PasswordHash); err == nil {
			encoded = user.PasswordHash
		}
	}
	err = s.hasher.Verify(ctx, encoded, password)
	if errors.Is(err, ErrRate) || errors.Is(err, ErrUnavailable) {
		return Principal{}, err
	}
	if nameErr != nil || err != nil || findErr != nil || !user.Enabled || encoded == s.dummy {
		return Principal{}, ErrInvalid
	}
	if _, err = s.store.Attempt(ctx, account, now, true); err != nil {
		return Principal{}, ErrUnavailable
	}
	return Principal{UserID: user.ID}, nil
}
func (s *Service) Bootstrap(ctx context.Context, username string, password []byte) (string, error) {
	if _, err := storage.NormalizeUsername(username); err != nil {
		return "", storage.ErrInvalid
	}
	hash, err := s.hasher.Hash(ctx, password)
	if err != nil {
		return "", err
	}
	return s.store.Bootstrap(ctx, username, hash)
}
