package auth

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func authFixture(t *testing.T) (*storage.SQLite, *Service) {
	t.Helper()
	s, err := storage.Open(context.Background(), storage.Options{Path: filepath.Join(t.TempDir(), "data", "hub.db"), BusyTimeout: time.Second, StartupTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	a, err := New(context.Background(), s, NewHasher())
	if err != nil {
		t.Fatal(err)
	}
	return s, a
}
func TestArgon2PolicySaltVerifyAndBounds(t *testing.T) {
	h := NewHasher()
	ctx := context.Background()
	password := []byte("Synthetic strong password 42")
	first, err := h.Hash(ctx, password)
	if err != nil {
		t.Fatal(err)
	}
	second, err := h.Hash(ctx, password)
	if err != nil || first == second {
		t.Fatal("salt reused", err)
	}
	if h.Verify(ctx, first, password) != nil || h.Verify(ctx, first, []byte("wrong")) == nil || NeedsRehash(first) {
		t.Fatal("verification")
	}
	for _, encoded := range []string{"bad", strings.Replace(first, "m=19456", "m=999999999", 1), strings.Replace(first, "v=19", "v=18", 1), strings.Replace(first, "t=2", "t=0", 1), first + "="} {
		if h.Verify(ctx, encoded, password) == nil {
			t.Fatal("bad PHC accepted")
		}
	}
	if _, err = h.Hash(ctx, []byte("short")); !errors.Is(err, ErrPassword) {
		t.Fatal(err)
	}
	if PasswordPolicy([]byte(strings.Repeat("я", 15))) != nil || PasswordPolicy([]byte(strings.Repeat("x", 1025))) == nil {
		t.Fatal("Unicode/size policy")
	}
	h.slots <- struct{}{}
	if _, err = h.Hash(ctx, password); !errors.Is(err, ErrRate) {
		t.Fatal("unbounded hash admission")
	}
	<-h.slots
}
func TestCredentialErrorsCooldownAndRestart(t *testing.T) {
	s, a := authFixture(t)
	ctx := context.Background()
	password := []byte("Synthetic strong password 42")
	id, err := a.Bootstrap(ctx, "Admin", password)
	if err != nil {
		t.Fatal(err)
	}
	p, err := a.Authenticate(ctx, "admin", password, "127.0.0.1")
	if err != nil || p.UserID != id {
		t.Fatal(p, err)
	}
	for _, name := range []string{"missing", "admin"} {
		if _, err = a.Authenticate(ctx, name, []byte("wrong"), "127.0.0.1"); !errors.Is(err, ErrInvalid) {
			t.Fatal(err)
		}
	}
	now := time.Now()
	a.clock = func() time.Time { return now }
	for i := 0; i < 4; i++ {
		_, err = a.Authenticate(ctx, "admin", []byte("wrong"), "127.0.0.1")
		if !errors.Is(err, ErrInvalid) {
			t.Fatal(i, err)
		}
	}
	if _, err = a.Authenticate(ctx, "admin", password, "127.0.0.1"); !errors.Is(err, ErrRate) {
		t.Fatal("account limit bypass", err)
	}
	again, err := New(ctx, s, NewHasher())
	if err != nil {
		t.Fatal(err)
	}
	again.clock = func() time.Time { return now }
	if _, err = again.Authenticate(ctx, "admin", password, "other-peer"); !errors.Is(err, ErrRate) {
		t.Fatal("restart cooldown bypass", err)
	}
	again.clock = func() time.Time { return now.Add(16 * time.Minute) }
	if _, err = again.Authenticate(ctx, "admin", password, "other-peer"); err != nil {
		t.Fatal(err)
	}
}
func TestDisabledUnknownMalformedAndLimiter(t *testing.T) {
	s, a := authFixture(t)
	ctx := context.Background()
	password := []byte("Synthetic strong password 42")
	id, err := a.Bootstrap(ctx, "invalid_username", password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = a.Authenticate(ctx, "bad/name", password, "peer"); !errors.Is(err, ErrInvalid) {
		t.Fatal("invalid username alias", err)
	}
	if err = s.Transaction(ctx, func(tx *sql.Tx) error { _, err := tx.Exec("UPDATE users SET enabled=0 WHERE id=?", id); return err }); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Authenticate(ctx, "invalid_username", password, "peer"); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	if err = s.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("UPDATE users SET enabled=1,password_hash='bad hash' WHERE id=?", id)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Authenticate(ctx, "invalid_username", password, "peer"); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	now := time.Now()
	for i := 0; i < 20; i++ {
		if !a.allowIP("new-peer", now) {
			t.Fatal(i)
		}
	}
	if a.allowIP("new-peer", now) {
		t.Fatal("IP limit bypass")
	}
	if !a.allowIP("new-peer", now.Add(time.Minute)) {
		t.Fatal("IP expiry")
	}
	for i := 0; i < 4096; i++ {
		a.ips[strings.Repeat("x", i%64)+time.Unix(int64(i), 0).String()] = ipWindow{start: now, count: 1}
	}
	if a.allowIP("full", now) {
		t.Fatal("unbounded limiter")
	}
}
func TestHasherRandomFailureAndCancellation(t *testing.T) {
	h := NewHasher()
	h.random = strings.NewReader("")
	if _, err := h.Hash(context.Background(), []byte("Synthetic strong password")); !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := NewHasher().Hash(ctx, []byte("Synthetic strong password")); err == nil {
		t.Fatal("cancellation ignored")
	}
}
func BenchmarkPasswordVerify(b *testing.B) {
	h := NewHasher()
	password := []byte("Synthetic strong password 42")
	encoded, err := h.Hash(context.Background(), password)
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if err := h.Verify(context.Background(), encoded, password); err != nil {
			b.Fatal(err)
		}
	}
}
