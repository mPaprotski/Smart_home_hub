package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mPaprotski/Smart_home_hub/internal/commands"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func accessFixture(t *testing.T) (*storage.SQLite, *Access, string) {
	t.Helper()
	s, service := authFixture(t)
	id, err := service.Bootstrap(context.Background(), "admin", []byte("Synthetic only long password 42"))
	if err != nil {
		t.Fatal(err)
	}
	return s, NewAccess(s, service), id
}
func TestSessionIdleAbsoluteCapsRestartAndFixation(t *testing.T) {
	s, a, user := accessFixture(t)
	ctx := context.Background()
	var clock atomic.Int64
	clock.Store(time.Now().Unix())
	a.Now = func() time.Time { return time.Unix(clock.Load(), 0) }
	raw, csrf, err := a.Issue(ctx, user)
	if err != nil || len(raw) != 64 || len(csrf) != 64 {
		t.Fatal(err)
	}
	c, err := a.Verify(ctx, raw, "", "peer")
	if err != nil || !ValidCSRF(c, csrf) || ValidCSRF(c, "wrong") {
		t.Fatal(err)
	}
	if data, _ := json.Marshal(c); strings.Contains(string(data), csrf) || strings.Contains(fmt.Sprintf("%#v", c), c.ID) {
		t.Fatal("credential format leak")
	}
	for i := 0; i < 5; i++ {
		clock.Add(1)
		fresh, _, err := a.Issue(ctx, user)
		if err != nil || fresh == raw {
			t.Fatal(err)
		}
	}
	if _, err = a.Verify(ctx, raw, "", "peer"); err == nil {
		t.Fatal("session cap failed")
	}
	raw, _, err = a.Issue(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	again := NewAccess(s, a.Credentials)
	again.Now = a.Now
	if _, err = again.Verify(ctx, raw, "", "peer"); err != nil {
		t.Fatal("restart session invalid", err)
	}
	clock.Add(1800)
	if _, err = a.Verify(ctx, raw, "", "peer"); err == nil {
		t.Fatal("idle expiry")
	}
	raw, _, _ = a.Issue(ctx, user)
	for i := 0; i < 25; i++ {
		clock.Add(1700)
		_, _ = a.Verify(ctx, raw, "", "peer")
	}
	clock.Add(1700)
	if _, err = a.Verify(ctx, raw, "", "peer"); err == nil {
		t.Fatal("absolute expiry extended")
	}
}
func TestTokenGrantAuditRevokeAndCredentialBeforeDedupe(t *testing.T) {
	s, a, user := accessFixture(t)
	ctx := context.Background()
	raw, _, err := a.Issue(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	c, err := a.Verify(ctx, raw, "", "peer")
	if err != nil {
		t.Fatal(err)
	}
	meta, token, err := a.CreateToken(ctx, c, "fixture", []string{"settings:write"}, []string{"user:" + user})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := s.ListTokens(ctx, c, a.Now())
	if err != nil || len(rows) != 1 {
		t.Fatal(err)
	}
	encoded, _ := json.Marshal(rows)
	if strings.Contains(string(encoded), strings.Split(token, ".")[1]) {
		t.Fatal("token revealed again")
	}
	tc, err := a.Verify(ctx, "", token, "peer")
	if err != nil || tc.ID != meta.ID {
		t.Fatal(err)
	}
	d := commands.New(s)
	p := commands.Principal{ID: tc.ID, UserID: user, Kind: "service", Scopes: tc.Scopes, Bindings: tc.Bindings, Credential: &tc}
	cmd := commands.Command{ID: "retry", Version: 1, Type: "SetUserSetting", Target: user, Key: "timezone", Value: "UTC", ExpiresAt: a.Now().Add(time.Minute)}
	if _, err = d.Execute(ctx, p, "EXTERNAL_API", cmd); err != nil {
		t.Fatal(err)
	}
	if err = a.RevokeToken(ctx, c, meta.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = d.Execute(ctx, p, "EXTERNAL_API", cmd); err == nil {
		t.Fatal("cached command after revoke")
	}
	if _, err = a.Verify(ctx, "", token, "peer"); err == nil {
		t.Fatal("revoked bearer")
	}
	for _, scopes := range [][]string{{"unknown"}, {"tokens:manage"}, {"lights:write"}} {
		if _, _, err = a.CreateToken(ctx, c, "bad", scopes, []string{"user:" + user}); err == nil {
			t.Fatal("overgrant", scopes)
		}
	}
	if _, _, err = a.CreateToken(ctx, c, "bad", []string{"settings:write"}, []string{"user:other"}); err == nil {
		t.Fatal("foreign binding")
	}
	audit, err := s.ListAudit(ctx, c, 0, a.Now())
	if err != nil || len(audit) < 4 {
		t.Fatal(audit, err)
	}
	for _, entry := range audit {
		if strings.Contains(fmt.Sprint(entry), token) || strings.Contains(fmt.Sprint(entry), csrfPlaceholder()) {
			t.Fatal("audit leak")
		}
	}
	if _, _, err = a.CreateToken(ctx, tc, "bad", []string{"settings:write"}, []string{"user:" + user}); err == nil {
		t.Fatal("machine token delegated")
	}
}
func csrfPlaceholder() string { return "Synthetic only long password 42" }
func TestSessionRevokeWatchRaceAndAuditFailure(t *testing.T) {
	s, a, user := accessFixture(t)
	ctx := context.Background()
	raw, _, err := a.Issue(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	c, err := a.Verify(ctx, raw, "", "peer")
	if err != nil {
		t.Fatal(err)
	}
	watch, cancel := context.WithCancel(ctx)
	remove, err := a.Watch(ctx, c, cancel)
	if err != nil {
		t.Fatal(err)
	}
	defer remove()
	if err = a.Revoke(ctx, c, true); err != nil {
		t.Fatal(err)
	}
	select {
	case <-watch.Done():
	case <-time.After(time.Second):
		t.Fatal("watch not canceled")
	}
	if _, err = a.Verify(ctx, raw, "", "peer"); err == nil {
		t.Fatal("logout bypass")
	}
	if _, err = a.Watch(ctx, c, func() {}); err == nil {
		t.Fatal("register-after-revoke race")
	}
	raw, _, _ = a.Issue(ctx, user)
	c, _ = a.Verify(ctx, raw, "", "peer")
	if err = s.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("CREATE TRIGGER deny_audit BEFORE INSERT ON security_audit BEGIN SELECT RAISE(ABORT,'fixture'); END")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, _, err = a.Issue(ctx, user); err == nil {
		t.Fatal("session without audit")
	}
	if err = a.Revoke(ctx, c, false); err == nil {
		t.Fatal("revoked without durable audit")
	}
	if _, err = a.Verify(ctx, raw, "", "peer"); err != nil {
		t.Fatal("failed revoke committed", err)
	}
	if _, _, err = a.CreateToken(ctx, c, "bad", []string{"settings:write"}, []string{"user:" + user}); err == nil {
		t.Fatal("token without audit")
	}
	if err = s.Transaction(ctx, func(tx *sql.Tx) error { _, err := tx.Exec("DROP TRIGGER deny_audit"); return err }); err != nil {
		t.Fatal(err)
	}
	if err = s.Transaction(ctx, func(tx *sql.Tx) error { _, err := tx.Exec("UPDATE users SET enabled=0 WHERE id=?", user); return err }); err != nil {
		t.Fatal(err)
	}
	if _, err = a.Verify(ctx, raw, "", "peer"); err == nil {
		t.Fatal("disabled owner")
	}
}
func TestCredentialLimitsAndSafeMalformedInputs(t *testing.T) {
	_, a, user := accessFixture(t)
	ctx := context.Background()
	raw, _, _ := a.Issue(ctx, user)
	c, _ := a.Verify(ctx, raw, "", "peer")
	for i := 0; i < 2; i++ {
		_, err := a.Watch(ctx, c, func() {})
		if err != nil {
			t.Fatal(err)
		}
	}
	if _, err := a.Watch(ctx, c, func() {}); !errors.Is(err, ErrRate) {
		t.Fatal(err)
	}
	if _, err := a.Verify(ctx, raw, "ambiguous", "peer"); err != storage.ErrInvalid {
		t.Fatal(err)
	}
	if _, err := a.Verify(ctx, "", strings.Repeat("x", 10000), "peer"); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	for i := 0; i < 16; i++ {
		if _, _, err := a.CreateToken(ctx, c, fmt.Sprintf("token-%d", i), []string{"settings:write"}, []string{"user:" + user}); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := a.CreateToken(ctx, c, "full", []string{"settings:write"}, []string{"user:" + user}); err != storage.Error("TOKEN_LIMIT") {
		t.Fatal(err)
	}
	a.Close()
	if a.Connections() != 0 {
		t.Fatal("watchers leaked")
	}
}

func TestTokenExpiryAndLiveRoleRemoval(t *testing.T) {
	s, a, user := accessFixture(t)
	ctx := context.Background()
	now := time.Now()
	a.Now = func() time.Time { return now }
	raw, _, err := a.Issue(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	c, err := a.Verify(ctx, raw, "", "peer")
	if err != nil {
		t.Fatal(err)
	}
	_, token, err := a.CreateToken(ctx, c, "writer", []string{"settings:write"}, []string{"user:" + user})
	if err != nil {
		t.Fatal(err)
	}
	tc, err := a.Verify(ctx, "", token, "peer")
	if err != nil {
		t.Fatal(err)
	}
	err = s.Transaction(ctx, func(tx *sql.Tx) error { _, e := tx.Exec("DELETE FROM user_roles WHERE user_id=?", user); return e })
	if err != nil {
		t.Fatal(err)
	}
	command := commands.Command{ID: "role-removed", Version: 1, Type: "SetUserSetting", Target: user, Key: "timezone", Value: "UTC", ExpiresAt: time.Now().Add(time.Minute)}
	_, err = commands.New(s).Execute(ctx, commands.Principal{ID: tc.ID, UserID: user, Kind: "service", Credential: &tc}, "EXTERNAL_API", command)
	if err == nil {
		t.Fatal("removed role still authorized")
	}
	now = now.Add(90 * 24 * time.Hour)
	if _, err = a.Verify(ctx, "", token, "peer"); err == nil {
		t.Fatal("expired token accepted")
	}
}

func TestVerifiedSessionRevokedBeforeAdmission(t *testing.T) {
	s, a, user := accessFixture(t)
	ctx := context.Background()
	raw, _, err := a.Issue(ctx, user)
	if err != nil {
		t.Fatal(err)
	}
	c, err := a.Verify(ctx, raw, "", "peer")
	if err != nil {
		t.Fatal(err)
	}
	if err = a.Revoke(ctx, c, false); err != nil {
		t.Fatal(err)
	}
	_, err = commands.New(s).Execute(ctx, commands.Principal{ID: user, UserID: user, Kind: "user", Credential: &c}, "WEB_UI", commands.Command{ID: "verified-before-revoke", Version: 1, Type: "SetUserSetting", Target: user, Key: "timezone", Value: "UTC", ExpiresAt: time.Now().Add(time.Minute)})
	if err == nil {
		t.Fatal("verified principal bypassed transaction recheck")
	}
	if _, err = s.GetSetting(ctx, user, "timezone"); !errors.Is(err, storage.ErrNotFound) {
		t.Fatal("mutation after revoke", err)
	}
}
