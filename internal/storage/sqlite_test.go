package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func testStore(t *testing.T) *SQLite {
	t.Helper()
	s, err := Open(context.Background(), Options{Path: filepath.Join(t.TempDir(), "data", "hub.db"), BusyTimeout: 50 * time.Millisecond, StartupTimeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}
func TestSQLitePersistenceConstraintsAndCAS(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	id, err := s.Bootstrap(ctx, "Admin", "synthetic-not-an-auth-fixture")
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.Diagnostics(ctx)
	if err != nil || d.JournalMode != "wal" || d.ForeignKeys != 1 || d.Synchronous != 2 || d.SchemaVersion != len(Migrations()) || !patchedVersion(d.SQLiteVersion) {
		t.Fatal(d, err)
	}
	t.Logf("SQLite %s WAL/FK/FULL schema=%d", d.SQLiteVersion, d.SchemaVersion)
	err = s.Transaction(ctx, func(tx *sql.Tx) error { _, err := SetSettingTx(ctx, tx, id, "timezone", "Europe/Minsk", 0); return err })
	if err != nil {
		t.Fatal(err)
	}
	err = s.Transaction(ctx, func(tx *sql.Tx) error { _, err := SetSettingTx(ctx, tx, id, "timezone", "UTC", 0); return err })
	if !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	err = s.Transaction(ctx, func(tx *sql.Tx) error { _, err := SetSettingTx(ctx, tx, "missing", "timezone", "UTC", 0); return err })
	if err == nil {
		t.Fatal("foreign key ignored")
	}
	err = s.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := SetSettingTx(ctx, tx, id, "timezone", "UTC", 1)
		if err != nil {
			return err
		}
		return ErrInvalid
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatal(err)
	}
	value, err := s.GetSetting(ctx, id, "timezone")
	if err != nil || value.Value != "Europe/Minsk" || value.Revision != 1 {
		t.Fatal(value, err)
	}
	var path string
	if err = s.db.QueryRow("SELECT file FROM pragma_database_list WHERE name='main'").Scan(&path); err != nil {
		t.Fatal(err)
	}
	if err = s.Close(); err != nil {
		t.Fatal(err)
	}
	again, err := Open(ctx, Options{Path: path, BusyTimeout: 50 * time.Millisecond, StartupTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer again.Close()
	value, err = again.GetSetting(ctx, id, "timezone")
	if err != nil || value.Revision != 1 || value.Value != "Europe/Minsk" {
		t.Fatal(value, err)
	}
	if _, err = again.Bootstrap(ctx, "second", "hash"); !errors.Is(err, ErrInitialized) {
		t.Fatal(err)
	}
}
func TestMigrationRollbackChecksumAndNewerSchema(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	bad := append(Migrations(), Migration{len(Migrations()) + 1, "CREATE TABLE should_rollback(id INTEGER); INSERT INTO missing_table VALUES(1);"})
	if err := Migrate(ctx, s.db, bad); err == nil {
		t.Fatal("failed step accepted")
	}
	var count int
	if s.db.QueryRow("SELECT count(*) FROM sqlite_master WHERE name='should_rollback'").Scan(&count) != nil || count != 0 {
		t.Fatal("half schema remains")
	}
	var version int
	_ = s.db.QueryRow("SELECT max(version) FROM schema_migrations").Scan(&version)
	if version != len(Migrations()) {
		t.Fatal(version)
	}
	upgrade := append(Migrations(), Migration{len(Migrations()) + 1, "CREATE TABLE upgrade_fixture(id INTEGER);"})
	if err := Migrate(ctx, s.db, upgrade); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, s.db, upgrade); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, s.db, Migrations()); err == nil {
		t.Fatal("newer schema accepted")
	}
	changed := append([]Migration(nil), upgrade...)
	changed[0].SQL += "\n"
	if err := Migrate(ctx, s.db, changed); err == nil {
		t.Fatal("checksum changed")
	}
}
func TestBusyCancellationAndConnectionPragmas(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	var path string
	_ = s.db.QueryRow("SELECT file FROM pragma_database_list WHERE name='main'").Scan(&path)
	other, err := Open(ctx, Options{Path: path, BusyTimeout: 50 * time.Millisecond, StartupTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	err = other.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO devices VALUES('blocked','fake','fake',1)")
		return err
	})
	if err != Error("STORAGE_BUSY") || time.Since(started) > time.Second {
		t.Fatal("busy not bounded", err)
	}
	_ = tx.Rollback()
	if err = other.Transaction(ctx, func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO devices VALUES('allowed','fake','fake',1)")
		return err
	}); err != nil {
		t.Fatal(err)
	}
	cancelCtx, cancel := context.WithCancel(ctx)
	cancel()
	if other.Transaction(cancelCtx, func(*sql.Tx) error { t.Fatal("canceled callback invoked"); return nil }) == nil {
		t.Fatal("canceled operation accepted")
	}
	s.db.SetMaxIdleConns(0)
	for i := 0; i < 3; i++ {
		d, err := s.Diagnostics(ctx)
		if err != nil || d.ForeignKeys != 1 || d.Synchronous != 2 {
			t.Fatal("new connection lost pragma", d, err)
		}
	}
}
func TestBootstrapConcurrentAndRollback(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	var mu sync.Mutex
	wins := 0
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Bootstrap(ctx, "admin", "synthetic-hash")
			if err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			} else if !errors.Is(err, ErrInitialized) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatal(wins)
	}
	if _, err := s.db.Exec("DELETE FROM users"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Bootstrap(ctx, "new", "hash"); !errors.Is(err, ErrInitialized) {
		t.Fatal("marker bypass", err)
	}
	rollback := testStore(t)
	_, err := rollback.db.Exec("CREATE TRIGGER abort_binding BEFORE INSERT ON user_roles BEGIN SELECT RAISE(ABORT,'synthetic failure'); END")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = rollback.Bootstrap(ctx, "admin", "synthetic-hash"); err == nil {
		t.Fatal("bootstrap ignored rollback")
	}
	var n int
	_ = rollback.db.QueryRow("SELECT count(*) FROM users").Scan(&n)
	if n != 0 {
		t.Fatal("partial user")
	}
	_ = rollback.db.QueryRow("SELECT count(*) FROM bootstrap_state").Scan(&n)
	if n != 0 {
		t.Fatal("partial marker")
	}
}
func TestInvalidPrivateStorageAndCorruption(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	if err := os.Chmod(dir, 0755); err != nil {
		t.Fatal(err)
	}
	_, err := Open(ctx, Options{Path: filepath.Join(dir, "x.db"), BusyTimeout: time.Millisecond, StartupTimeout: time.Second})
	if err == nil {
		t.Fatal("public directory accepted")
	}
	if strings.Contains(err.Error(), dir) {
		t.Fatal("path leak")
	}
	dir = t.TempDir()
	if err = os.Chmod(dir, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "bad.db")
	if err = os.WriteFile(path, []byte("not SQLite"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err = Open(ctx, Options{Path: path, BusyTimeout: time.Millisecond, StartupTimeout: time.Second}); err == nil {
		t.Fatal("corrupt db accepted")
	}
	if _, err = Open(ctx, Options{Path: ":memory:", BusyTimeout: time.Millisecond, StartupTimeout: time.Second}); err == nil {
		t.Fatal("memory labeled persistent")
	}
}
func TestAccountAttemptsSurviveReopen(t *testing.T) {
	s := testStore(t)
	now := time.Unix(1000, 0)
	for i := 0; i < 5; i++ {
		ok, err := s.Attempt(context.Background(), "keyed-digest", now, false)
		if err != nil || !ok {
			t.Fatal(ok, err)
		}
	}
	if ok, _ := s.Attempt(context.Background(), "keyed-digest", now, false); ok {
		t.Fatal("limit bypass")
	}
	if ok, err := s.Attempt(context.Background(), "keyed-digest", now.Add(15*time.Minute), false); err != nil || !ok {
		t.Fatal("expiry failed", err)
	}
}

func TestBoundedDiskFullRollback(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	if _, err := s.db.Exec("CREATE TABLE full_fixture(data BLOB)"); err != nil {
		t.Fatal(err)
	}
	var pages int
	if err := s.db.QueryRow("PRAGMA page_count").Scan(&pages); err != nil {
		t.Fatal(err)
	}
	if _, err := s.db.Exec(fmt.Sprintf("PRAGMA max_page_count=%d", pages+1)); err != nil {
		t.Fatal(err)
	}
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO full_fixture VALUES(zeroblob(1048576))"); err != nil {
			return err
		}
		return nil
	})
	if err != Error("STORAGE_FULL") {
		t.Fatal("bounded SQLITE_FULL injection failed", err)
	}
	var count int
	if err = s.db.QueryRow("SELECT count(*) FROM full_fixture").Scan(&count); err != nil || count != 0 {
		t.Fatal("disk full left partial write", count, err)
	}
}
func TestStorageProcessHelper(t *testing.T) {
	if os.Getenv("SHH_STORAGE_HELPER") != "1" {
		return
	}
	s, err := Open(context.Background(), Options{Path: os.Getenv("SHH_STORAGE_PATH"), BusyTimeout: 100 * time.Millisecond, StartupTimeout: time.Second})
	if err != nil {
		os.Exit(20)
	}
	u, err := s.FindUser(context.Background(), "admin")
	if errors.Is(err, ErrNotFound) {
		if _, err = s.Bootstrap(context.Background(), "admin", "synthetic-process-fixture"); err != nil {
			os.Exit(21)
		}
		u, err = s.FindUser(context.Background(), "admin")
	}
	if err != nil {
		os.Exit(22)
	}
	if os.Getenv("SHH_STORAGE_ACTION") == "commit" {
		err = s.Transaction(context.Background(), func(tx *sql.Tx) error {
			_, err := SetSettingTx(context.Background(), tx, u.ID, "timezone", "Europe/Minsk", 0)
			return err
		})
		if err != nil {
			os.Exit(23)
		}
		os.Exit(0)
	}
	_ = s.Transaction(context.Background(), func(tx *sql.Tx) error {
		if _, err := SetSettingTx(context.Background(), tx, u.ID, "timezone", "UTC", 1); err != nil {
			os.Exit(24)
		}
		os.Exit(17)
		return nil
	})
	os.Exit(25)
}
func TestActualProcessRestartAndUncommittedRecovery(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data", "hub.db")
	run := func(action string, want int) {
		cmd := exec.Command(os.Args[0], "-test.run=^TestStorageProcessHelper$")
		cmd.Env = append(os.Environ(), "SHH_STORAGE_HELPER=1", "SHH_STORAGE_PATH="+path, "SHH_STORAGE_ACTION="+action)
		err := cmd.Run()
		code := 0
		if err != nil {
			var e *exec.ExitError
			if !errors.As(err, &e) {
				t.Fatal(err)
			}
			code = e.ExitCode()
		}
		if code != want {
			t.Fatal(action, code)
		}
	}
	run("commit", 0)
	run("uncommitted", 17)
	s, err := Open(context.Background(), Options{Path: path, BusyTimeout: 100 * time.Millisecond, StartupTimeout: time.Second})
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	u, err := s.FindUser(context.Background(), "admin")
	if err != nil {
		t.Fatal(err)
	}
	setting, err := s.GetSetting(context.Background(), u.ID, "timezone")
	if err != nil || setting.Value != "Europe/Minsk" || setting.Revision != 1 {
		t.Fatal("process recovery", setting, err)
	}
}

func TestZoneDesiredStateAndPrivacySeparation(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	z := Zone{ID: "monitor.backlight", Mode: "STATIC", Brightness: .2}
	if err := s.Transaction(ctx, func(tx *sql.Tx) error { _, err := SetZoneTx(ctx, tx, z, 0); return err }); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetZone(ctx, z.ID)
	if err != nil || got.Revision != 1 || got.Brightness != .2 {
		t.Fatal(got, err)
	}
	z.Brightness = math.NaN()
	if err = s.Transaction(ctx, func(tx *sql.Tx) error { _, err := SetZoneTx(ctx, tx, z, 1); return err }); !errors.Is(err, ErrInvalid) {
		t.Fatal("NaN desired state", err)
	}
	rows, err := s.db.Query("SELECT name FROM sqlite_master WHERE type='table'")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if rows.Scan(&name) != nil {
			t.Fatal("scan")
		}
		if name == "events" || name == "observations" || name == "frames" {
			t.Fatal("forbidden history table", name)
		}
	}
}

func TestReadOnlyStorageErrorClassification(t *testing.T) {
	s := testStore(t)
	if _, err := s.db.Exec("PRAGMA query_only=ON"); err != nil {
		t.Fatal(err)
	}
	err := s.Transaction(context.Background(), func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO devices VALUES('read-only','fake','fake',1)")
		return err
	})
	if err != Error("STORAGE_READ_ONLY") {
		t.Fatal(err)
	}
}

func TestExistingReadOnlyDatabaseRejectsWriteProbe(t *testing.T) {
	s := testStore(t)
	var path string
	if err := s.db.QueryRow("SELECT file FROM pragma_database_list WHERE name='main'").Scan(&path); err != nil {
		t.Fatal(err)
	}
	u := url.URL{Scheme: "file", Path: path}
	q := url.Values{}
	q.Set("mode", "ro")
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "busy_timeout(50)")
	q.Set("_txlock", "immediate")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	readonly := &SQLite{db: db}
	if err = readonly.Ping(context.Background()); err != nil {
		t.Fatal("fixture must remain readable", err)
	}
	err = readonly.Transaction(context.Background(), func(tx *sql.Tx) error {
		_, err := tx.Exec("UPDATE schema_migrations SET applied_at=applied_at WHERE version=1")
		return err
	})
	if err != Error("STORAGE_READ_ONLY") {
		t.Fatal("write probe accepted readonly database", err)
	}
}
