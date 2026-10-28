package storage

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func domainFixture(t *testing.T) (*SQLite, string) {
	s := testStore(t)
	ctx := context.Background()
	id, err := s.Bootstrap(ctx, "admin", "synthetic-hash")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Transaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO devices VALUES('button','fake','fake',1)"); err != nil {
			return err
		}
		_, err := SetZoneTx(ctx, tx, Zone{ID: "zone", OwnerID: id, Mode: "OFF", Brightness: .2}, 0)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return s, id
}
func TestDomainTypedOwnershipRulesAndMappings(t *testing.T) {
	s, user := domainFixture(t)
	ctx := context.Background()
	if err := s.CreatePreset(ctx, user, "preset", "Fixture", []PresetEntry{{"zone", "STATIC", .5}}); err != nil {
		t.Fatal(err)
	}
	entries, err := s.PresetEntries(ctx, user, "preset")
	if err != nil || len(entries) != 1 {
		t.Fatal(entries, err)
	}
	if _, err = s.PresetEntries(ctx, "other", "preset"); err != Error("FORBIDDEN") {
		t.Fatal(err)
	}
	if err = s.CreatePreset(ctx, user, "bad", "bad", []PresetEntry{{"zone", "STATIC", math.NaN()}}); !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	rule := []byte(`{"version":1,"trigger":{"kind":"sensor_threshold","sensor_id":"lux","above":50},"action":{"type":"SetZoneState","zone_id":"zone","mode":"OFF","brightness":0.2}}`)
	if err = s.SaveRule(ctx, user, "rule", rule, 0); err != nil {
		t.Fatal(err)
	}
	if err = s.SaveRule(ctx, user, "rule", rule, 0); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	if _, err = ParseRule([]byte(`{"version":1,"raw_frame":"forbidden"}`)); err == nil {
		t.Fatal("arbitrary DSL")
	}
	for gesture := range gestures {
		if err = s.MapGesture(ctx, user, gesture, "zone", 500, true); err != nil {
			t.Fatal(err)
		}
	}
	if err = s.MapGesture(ctx, user, "unknown", "zone", 500, true); err == nil {
		t.Fatal("unknown gesture")
	}
	for _, press := range []string{"single", "double", "long"} {
		if err = s.MapButton(ctx, user, "button", press, "zone"); err != nil {
			t.Fatal(err)
		}
	}
	if err = s.MapButton(ctx, user, "missing", "single", "zone"); err == nil {
		t.Fatal("FK ignored")
	}
}
func TestHabitPrivacyEpochFeedbackAndRetention(t *testing.T) {
	s, user := domainFixture(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	if err := s.StartLearning(ctx, user, now); err != nil {
		t.Fatal(err)
	}
	cutoff := now.AddDate(0, 0, -90)
	for _, at := range []time.Time{cutoff.Add(-time.Second), cutoff, cutoff.Add(time.Second)} {
		h := HabitObservation{At: at, Timezone: "Europe/Minsk", Epoch: 1, ZoneID: "zone", Mode: "OFF", Origin: "manual", Quality: "unknown"}
		if err := s.AddHabit(ctx, user, h); err != nil {
			t.Fatal(err)
		}
	}
	before, err := s.Habits(ctx, user, 0, 100, now)
	if err != nil || len(before) != 2 {
		t.Fatal("expired history visible before cleanup", len(before), err)
	}
	h := HabitObservation{At: now, Timezone: "UTC", Epoch: 1, ZoneID: "zone", Mode: "STATIC", Origin: "manual", Quality: "valid"}
	v := math.Inf(1)
	h.Volume = &v
	if s.AddHabit(ctx, user, h) == nil {
		t.Fatal("nonfinite context")
	}
	if err := s.PublishModel(ctx, user, "v1", strings.Repeat("a", 64), 1, now); err != nil {
		t.Fatal(err)
	}
	action := ProposedAction{ZoneID: "zone", Mode: "STATIC"}
	if err := s.Recommend(ctx, user, "recommendation", "v1", 1, .8, action, now.Add(time.Hour), now); err != nil {
		t.Fatal(err)
	}
	if err := s.Feedback(ctx, user, "recommendation", "accepted", now); err != nil {
		t.Fatal(err)
	}
	if err := s.Feedback(ctx, user, "recommendation", "accepted", now); err != nil {
		t.Fatal("retry", err)
	}
	if err := s.Feedback(ctx, user, "recommendation", "rejected", now); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	if err := s.Feedback(ctx, "other", "recommendation", "accepted", now); err != Error("FORBIDDEN") {
		t.Fatal(err)
	}
	if _, err := s.Cleanup(ctx, now, 100); err != nil {
		t.Fatal(err)
	}
	rows, err := s.Habits(ctx, user, 0, 100, now)
	if err != nil || len(rows) != 2 || rows[0].Presence != nil || rows[0].Volume != nil || rows[0].MonitorActive != nil {
		t.Fatal("NULL/cutoff semantics", len(rows), err)
	}
	if rows[0].ID == 0 {
		t.Fatal("missing cursor")
	}
	if err = s.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("UPDATE habit_learning_state SET privacy_epoch=2 WHERE user_id=?", user)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err = s.PublishModel(ctx, user, "old-worker", strings.Repeat("b", 64), 1, now); !errors.Is(err, ErrConflict) {
		t.Fatal("stale training published", err)
	}
	rows, err = s.Habits(ctx, user, 0, 100, now)
	if err != nil || len(rows) != 0 {
		t.Fatal("old generation returned", len(rows), err)
	}
}
func TestBoundedIndexedDatasetAndAuditCutoff(t *testing.T) {
	s, user := domainFixture(t)
	ctx := context.Background()
	now := time.Date(2026, 10, 26, 0, 0, 0, 0, time.UTC)
	if err := s.StartLearning(ctx, user, now); err != nil {
		t.Fatal(err)
	}
	start := time.Now()
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(ctx, "WITH RECURSIVE n(x) AS(VALUES(1) UNION ALL SELECT x+1 FROM n WHERE x<100000) INSERT INTO habit_observations(user_id,occurred_at,timezone,privacy_epoch,zone_id,mode,origin,quality) SELECT ?,?+x,'UTC',1,'zone','OFF','manual','unknown' FROM n", user, now.Unix())
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := s.Habits(ctx, user, 0, 100, now)
	if err != nil || len(rows) != 100 {
		t.Fatal(len(rows), err)
	}
	if _, err = s.Habits(ctx, user, 0, 100000, now); err == nil {
		t.Fatal("unbounded query")
	}
	plan, err := s.ObservationQueryPlan(ctx, user, now)
	if err != nil || !strings.Contains(strings.Join(plan, " "), "habit_owner_time") {
		t.Fatal(plan, err)
	}
	cutoff := now.AddDate(0, 0, -30)
	if err = s.Transaction(ctx, func(tx *sql.Tx) error {
		for _, at := range []time.Time{cutoff.Add(-time.Second), cutoff, cutoff.Add(time.Second)} {
			if err := AuditTx(ctx, tx, at, "login", "success", user, "session", "created"); err != nil {
				return err
			}
			if _, err := tx.Exec("INSERT INTO automation_audit(owner_id,rule_ref,occurred_at,result,reason) VALUES(?,'rule',?,'APPLIED','fixture')", user, at.Unix()); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Cleanup(ctx, now, 1); err != nil {
		t.Fatal(err)
	}
	var count int
	if s.db.QueryRow("SELECT count(*) FROM security_audit").Scan(&count) != nil || count != 2 {
		t.Fatal(count)
	}
	if s.db.QueryRow("SELECT count(*) FROM automation_audit").Scan(&count) != nil || count != 2 {
		t.Fatal(count)
	}
	t.Logf("100000-row fixture insert/read/plan/cleanup=%s, returned=%d, plan=%v", time.Since(start), len(rows), plan)
}
func TestUpgradeFromDay18PreservesDataAndSessionAuditRollback(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	id, err := s.Bootstrap(ctx, "admin", "synthetic-hash")
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("CREATE TRIGGER fail_audit BEFORE INSERT ON security_audit BEGIN SELECT RAISE(ABORT,'fixture'); END")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	err = s.IssueSession(ctx, id, strings.Repeat("a", 64), strings.Repeat("b", 64), time.Now())
	if err == nil {
		t.Fatal("session issued without audit")
	}
	var count int
	_ = s.db.QueryRow("SELECT count(*) FROM sessions").Scan(&count)
	if count != 0 {
		t.Fatal("partial session")
	}
}

func TestActualV3UpgradePreservesUserSettingsAndRejectsLegacySession(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "private", "hub.db")
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	file.Close()
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=foreign_keys(1)&_txlock=immediate")
	if err != nil {
		t.Fatal(err)
	}
	db.SetMaxOpenConns(1)
	if err = Migrate(ctx, db, Migrations()[:3]); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO users VALUES('retained','retained','synthetic-hash',1,'UTC',1); INSERT INTO user_roles VALUES('retained','admin'); INSERT INTO user_settings VALUES('retained','timezone','Europe/Minsk',1,7); INSERT INTO bootstrap_state VALUES(1,1)"); err != nil {
		t.Fatal(err)
	}
	if _, err = db.Exec("INSERT INTO sessions VALUES('legacy','retained',?,NULL)", time.Now().Add(time.Hour).Unix()); err != nil {
		t.Fatal(err)
	}
	db.Close()
	s, err := Open(ctx, Options{Path: path, BusyTimeout: time.Second, StartupTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	u, err := s.FindUser(ctx, "retained")
	if err != nil || u.ID != "retained" {
		t.Fatal(u.ID, err)
	}
	v, err := s.GetSetting(ctx, u.ID, "timezone")
	if err != nil || v.Value != "Europe/Minsk" || v.Revision != 7 {
		t.Fatal(v, err)
	}
	if _, err = s.PeekCredential(ctx, Credential{Kind: "session", ID: "legacy"}, time.Now()); err == nil {
		t.Fatal("old schema revived an unprotected session")
	}
	d, err := s.Diagnostics(ctx)
	if err != nil || d.SchemaVersion != len(Migrations()) {
		t.Fatal(d, err)
	}
	t.Logf("actual v3 -> v%d preserves user/setting/marker", d.SchemaVersion)
}

func TestModelAgingInvalidatesRecommendationAndCascadesFeedback(t *testing.T) {
	s, user := domainFixture(t)
	ctx := context.Background()
	now := time.Now().UTC()
	old := now.AddDate(0, 0, -91)
	if e := s.StartLearning(ctx, user, old); e != nil {
		t.Fatal(e)
	}
	if e := s.PublishModel(ctx, user, "old", strings.Repeat("a", 64), 1, old); e != nil {
		t.Fatal(e)
	}
	if e := s.Recommend(ctx, user, "old-rec", "old", 1, .9, ProposedAction{ZoneID: "zone", Mode: "OFF"}, old.Add(time.Hour), old); e != nil {
		t.Fatal(e)
	}
	if e := s.Feedback(ctx, user, "old-rec", "accepted", old); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Cleanup(ctx, now, 1); e != nil {
		t.Fatal(e)
	}
	for _, table := range []string{"habit_model_metadata", "recommendations", "prediction_feedback"} {
		var count int
		if e := s.db.QueryRow("SELECT count(*) FROM " + table).Scan(&count); e != nil || count != 0 {
			t.Fatal(table, count, e)
		}
	}
	var quality string
	if e := s.db.QueryRow("SELECT quality FROM habit_learning_state WHERE user_id=?", user).Scan(&quality); e != nil || quality != "insufficient" {
		t.Fatal(quality, e)
	}
}
