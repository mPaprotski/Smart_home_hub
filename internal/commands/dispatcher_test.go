package commands

import (
	"context"
	"database/sql"
	"errors"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func fixture(t *testing.T) (*storage.SQLite, *Dispatcher, Principal) {
	t.Helper()
	s, err := storage.Open(context.Background(), storage.Options{Path: filepath.Join(t.TempDir(), "data", "hub.db"), BusyTimeout: time.Second, StartupTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	id, err := s.Bootstrap(context.Background(), "admin", "synthetic-hash-not-auth")
	if err != nil {
		t.Fatal(err)
	}
	return s, New(s), Principal{ID: id, UserID: id, Kind: "user"}
}
func request(p Principal) Command {
	return Command{ID: "one", Version: 1, Type: "SetUserSetting", Target: p.UserID, Key: "timezone", Value: "Europe/Minsk", ExpiresAt: time.Now().Add(time.Minute)}
}
func TestPersistentCommandIdempotencyAndReauthorization(t *testing.T) {
	s, d, p := fixture(t)
	ctx := context.Background()
	c := request(p)
	out, err := d.Execute(ctx, p, "LOCAL_CLI", c)
	if err != nil || out.Status != "APPLIED" || out.Revision != 1 {
		t.Fatal(out, err)
	}
	out, err = d.Execute(ctx, p, "LOCAL_CLI", c)
	if err != nil || out.Revision != 1 {
		t.Fatal(out, err)
	}
	c.Value = "UTC"
	if _, err = d.Execute(ctx, p, "LOCAL_CLI", c); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	c.Value = "Europe/Minsk"
	if err = s.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("UPDATE users SET enabled=0 WHERE id=?", p.UserID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = d.Execute(ctx, p, "LOCAL_CLI", c); err == nil {
		t.Fatal("cached result bypassed revoked rights")
	}
}
func TestDeniedInvalidOwnershipAndBindings(t *testing.T) {
	s, d, p := fixture(t)
	ctx := context.Background()
	base := request(p)
	for _, change := range []func(*Command){func(c *Command) { c.Target = "someone_else" }, func(c *Command) { c.Value = "unknown/timezone" }, func(c *Command) { c.Type = "SetZoneState" }, func(c *Command) { c.Version = 2 }, func(c *Command) { c.ExpiresAt = time.Unix(1, 0) }} {
		c := base
		change(&c)
		if out, err := d.Execute(ctx, p, "WEB_UI", c); err == nil || out.Status != "" {
			t.Fatal("invalid accepted", out, err)
		}
	}
	for _, scopes := range [][]string{{"settings:write", "unknown"}, {"settings:read"}} {
		service := Principal{ID: "service-1", UserID: p.UserID, Kind: "service", Scopes: scopes, Bindings: []string{"user:" + p.UserID}}
		if _, err := d.Execute(ctx, service, "EXTERNAL_API", base); err == nil {
			t.Fatal("scope bypass")
		}
	}
	service := Principal{ID: "service-1", UserID: p.UserID, Kind: "service", Scopes: []string{"settings:write"}, Bindings: []string{"user:other"}}
	if _, err := d.Execute(ctx, service, "EXTERNAL_API", base); err == nil {
		t.Fatal("binding bypass")
	}
	if _, err := d.Execute(ctx, p, "FORGED", base); err == nil {
		t.Fatal("source accepted")
	}
	if _, err := s.GetSetting(ctx, p.UserID, "timezone"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatal("denied mutated state", err)
	}
}
func TestConcurrentCASAndDuplicates(t *testing.T) {
	s, d, p := fixture(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c := request(p)
			if i%2 == 1 {
				c.ID = "two"
			}
			out, err := d.Execute(ctx, p, "WEB_UI", c)
			if err == nil {
				if out.Revision != 1 {
					t.Error(out)
				}
				mu.Lock()
				wins++
				mu.Unlock()
			} else if !errors.Is(err, storage.ErrConflict) {
				t.Error(err)
			}
		}(i)
	}
	wg.Wait()
	if wins != 4 {
		t.Fatal(wins)
	}
	v, err := s.GetSetting(ctx, p.UserID, "timezone")
	if err != nil || v.Revision != 1 {
		t.Fatal(v, err)
	}
}
func TestCommitFailureRollsBackAndNoFalseApplied(t *testing.T) {
	s, d, p := fixture(t)
	ctx := context.Background()
	if err := s.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("CREATE TRIGGER fail_record BEFORE INSERT ON command_records BEGIN SELECT RAISE(ABORT,'failure'); END")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	out, err := d.Execute(ctx, p, "WEB_UI", request(p))
	if err == nil || out.Status != "" {
		t.Fatal(out, err)
	}
	if _, err = s.GetSetting(ctx, p.UserID, "timezone"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatal("partial update", err)
	}
}
func TestLightingArbitrationClockAndRetry(t *testing.T) {
	s, _, p := fixture(t)
	ctx := context.Background()
	now := time.Unix(1000, 0)
	call := func(source, id, path string, at time.Time) error {
		return s.Transaction(ctx, func(tx *sql.Tx) error {
			return AdmitLighting(ctx, tx, Intent{PrincipalID: p.ID, CommandID: id, Source: source, Paths: []string{path}}, at)
		})
	}
	if err := call("WEB_UI", "manual", "lighting:zone:brightness", now); err != nil {
		t.Fatal(err)
	}
	if err := call("PREDICTION", "automatic", "lighting:zone:brightness", now.Add(time.Second)); !errors.Is(err, ErrOverride) {
		t.Fatal(err)
	}
	if err := call("AUTOMATION", "independent", "lighting:other:mode", now); err != nil {
		t.Fatal(err)
	}
	if err := call("WEB_UI", "manual", "lighting:zone:brightness", now.Add(800*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := call("AUTOMATION", "expired", "lighting:zone:brightness", now.Add(900*time.Second)); err != nil {
		t.Fatal("retry extended lease", err)
	}
	if err := call("WEB_UI", "master-off-fixture", "lighting:*", now.Add(901*time.Second)); err != nil {
		t.Fatal(err)
	}
	if err := call("PREDICTION", "blocked", "lighting:other:mode", now.Add(902*time.Second)); !errors.Is(err, ErrOverride) {
		t.Fatal(err)
	}
}

func TestRealSecondUserAndServiceIdentityIsolation(t *testing.T) {
	s, d, p := fixture(t)
	ctx := context.Background()
	if err := s.Transaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO users VALUES('other','other','synthetic',1,'UTC',0)"); err != nil {
			return storage.ErrStorage
		}
		if _, err := tx.Exec("INSERT INTO user_roles VALUES('other','admin')"); err != nil {
			return storage.ErrStorage
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	other := Principal{ID: "other", UserID: "other", Kind: "user"}
	c := request(p)
	if _, err := d.Execute(ctx, other, "WEB_UI", c); err == nil {
		t.Fatal("cross user mutation")
	}
	c.Target = "other"
	if _, err := d.Execute(ctx, other, "WEB_UI", c); err != nil {
		t.Fatal(err)
	}
	service := Principal{ID: "service-1", UserID: p.UserID, Kind: "service", Scopes: []string{"settings:write"}, Bindings: []string{"user:" + p.UserID}}
	c = request(p)
	if _, err := d.Execute(ctx, service, "EXTERNAL_API", c); err != nil {
		t.Fatal(err)
	}
	service.ID = "service-2"
	if _, err := d.Execute(ctx, service, "EXTERNAL_API", c); !errors.Is(err, storage.ErrConflict) {
		t.Fatal("service reused another identity cache", err)
	}
	if err := s.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM user_roles WHERE user_id=?", p.UserID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	service.ID = "service-1"
	if _, err := d.Execute(ctx, service, "EXTERNAL_API", c); err == nil {
		t.Fatal("role revoke bypassed")
	}
}
