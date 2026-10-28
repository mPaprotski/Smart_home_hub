package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"embed"
	"encoding/hex"
	"errors"
	"fmt"
	sqliteDriver "modernc.org/sqlite"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Errors deliberately contain no raw SQL, paths, values or driver details.
type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrStorage     Error = "STORAGE_UNAVAILABLE"
	ErrConflict    Error = "REVISION_CONFLICT"
	ErrNotFound    Error = "NOT_FOUND"
	ErrInitialized Error = "ALREADY_INITIALIZED"
	ErrInvalid     Error = "INVALID_VALUE"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

type Migration struct {
	Version int
	SQL     string
}

func Migrations() []Migration {
	var out []Migration
	for i, name := range []string{"001_core.sql", "002_commands.sql", "003_auth.sql", "004_sessions.sql", "005_tokens_audit.sql", "006_domains.sql", "007_retention_indexes.sql"} {
		b, _ := migrationFiles.ReadFile("migrations/" + name)
		out = append(out, Migration{i + 1, string(b)})
	}
	return out
}

type Options struct {
	Path           string
	BusyTimeout    time.Duration
	StartupTimeout time.Duration
}
type SQLite struct{ db *sql.DB }

func Open(ctx context.Context, o Options) (*SQLite, error) {
	if o.Path == "" || o.Path == ":memory:" || o.BusyTimeout < time.Millisecond || o.BusyTimeout > 5*time.Second || o.StartupTimeout < time.Millisecond || o.StartupTimeout > time.Minute {
		return nil, ErrInvalid
	}
	path, err := filepath.Abs(o.Path)
	if err != nil {
		return nil, ErrStorage
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return nil, ErrStorage
	}
	// A database and its sidecars must not be writable/readable by other OS users.
	info, err := os.Stat(filepath.Dir(path))
	if err != nil || info.Mode().Perm()&0077 != 0 {
		return nil, Error("DATA_DIRECTORY_NOT_PRIVATE")
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0600)
	if err == nil {
		_ = f.Close()
	} else if !errors.Is(err, os.ErrExist) {
		return nil, ErrStorage
	}
	info, err = os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0077 != 0 {
		return nil, Error("DATABASE_NOT_PRIVATE")
	}
	u := url.URL{Scheme: "file", Path: path}
	q := url.Values{}
	q.Add("_pragma", "foreign_keys(1)")
	q.Add("_pragma", "busy_timeout("+strconv.FormatInt(o.BusyTimeout.Milliseconds(), 10)+")")
	q.Add("_pragma", "synchronous(FULL)")
	q.Set("_txlock", "immediate")
	u.RawQuery = q.Encode()
	db, err := sql.Open("sqlite", u.String())
	if err != nil {
		return nil, ErrStorage
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	s := &SQLite{db: db}
	success := false
	defer func() {
		if !success {
			_ = db.Close()
		}
	}()
	startup, cancel := context.WithTimeout(ctx, o.StartupTimeout)
	defer cancel()
	var mode, version string
	var fk, syncMode int
	if err := db.QueryRowContext(startup, "PRAGMA journal_mode=WAL").Scan(&mode); err != nil {
		return nil, safeSQL(err)
	}
	if mode != "wal" || db.QueryRowContext(startup, "PRAGMA foreign_keys").Scan(&fk) != nil || fk != 1 || db.QueryRowContext(startup, "PRAGMA synchronous").Scan(&syncMode) != nil || syncMode != 2 {
		return nil, ErrStorage
	}
	if db.QueryRowContext(startup, "SELECT sqlite_version()").Scan(&version) != nil || !patchedVersion(version) {
		return nil, Error("SQLITE_VERSION_UNSUPPORTED")
	}
	if err = Migrate(startup, db, Migrations()); err != nil {
		return nil, err
	}
	// A fully migrated WAL database can be opened read-only when sidecars exist.
	// Acquire an actual write transaction, without changing migration metadata.
	if err = s.Transaction(startup, func(tx *sql.Tx) error {
		_, err := tx.ExecContext(startup, "UPDATE schema_migrations SET applied_at=applied_at WHERE version=1")
		return err
	}); err != nil {
		return nil, err
	}
	success = true
	return s, nil
}
func patchedVersion(v string) bool {
	var major, minor, patch int
	if _, err := fmt.Sscanf(v, "%d.%d.%d", &major, &minor, &patch); err != nil {
		return false
	}
	return major == 3 && (minor > 51 || minor == 51 && patch >= 3 || minor == 50 && patch >= 7 || minor == 44 && patch >= 6)
}

// Each step and its version record commit atomically under an immediate write lock.
func Migrate(ctx context.Context, db *sql.DB, migrations []Migration) error {
	for i, m := range migrations {
		if m.Version != i+1 {
			return Error("MIGRATION_ORDER")
		}
	}
	for target := 0; target <= len(migrations); target++ {
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return ErrStorage
		}
		run := func() error {
			if _, err := tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, checksum TEXT NOT NULL, applied_at INTEGER NOT NULL)"); err != nil {
				return ErrStorage
			}
			rows, err := tx.QueryContext(ctx, "SELECT version,checksum FROM schema_migrations ORDER BY version")
			if err != nil {
				return ErrStorage
			}
			count := 0
			for rows.Next() {
				var version int
				var checksum string
				if rows.Scan(&version, &checksum) != nil {
					rows.Close()
					return ErrStorage
				}
				count++
				if version != count || version > len(migrations) {
					rows.Close()
					return Error("SCHEMA_VERSION_UNSUPPORTED")
				}
				sum := sha256.Sum256([]byte(migrations[version-1].SQL))
				if checksum != hex.EncodeToString(sum[:]) {
					rows.Close()
					return Error("MIGRATION_CHECKSUM")
				}
			}
			err = rows.Err()
			rows.Close()
			if err != nil {
				return ErrStorage
			}
			if target == len(migrations) || count >= target+1 {
				return nil
			}
			if count != target {
				return Error("MIGRATION_ORDER")
			}
			m := migrations[target]
			if _, err = tx.ExecContext(ctx, m.SQL); err != nil {
				return Error("MIGRATION_FAILED")
			}
			sum := sha256.Sum256([]byte(m.SQL))
			if _, err = tx.ExecContext(ctx, "INSERT INTO schema_migrations VALUES(?,?,?)", m.Version, hex.EncodeToString(sum[:]), time.Now().UTC().Unix()); err != nil {
				return ErrStorage
			}
			return nil
		}()
		if run != nil {
			_ = tx.Rollback()
			return run
		}
		if tx.Commit() != nil {
			return ErrStorage
		}
	}
	return nil
}
func (s *SQLite) Close() error {
	if s.db.Close() != nil {
		return ErrStorage
	}
	return nil
}
func (s *SQLite) Ping(ctx context.Context) error {
	if s.db.PingContext(ctx) != nil {
		return ErrStorage
	}
	return nil
}

// Transaction is a trusted internal unit-of-work; callers must not perform network,
// hashing or hardware I/O while holding it. Public errors are always sanitized.
func (s *SQLite) Transaction(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return safeSQL(err)
	}
	defer tx.Rollback()
	if err = fn(tx); err != nil {
		var driver *sqliteDriver.Error
		if errors.As(err, &driver) {
			return safeSQL(err)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return safeSQL(err)
	}
	return nil
}

type Diagnostics struct {
	SQLiteVersion string `json:"sqlite_version"`
	JournalMode   string `json:"journal_mode"`
	ForeignKeys   int    `json:"foreign_keys"`
	Synchronous   int    `json:"synchronous"`
	SchemaVersion int    `json:"schema_version"`
}

func (s *SQLite) Diagnostics(ctx context.Context) (Diagnostics, error) {
	var d Diagnostics
	for _, p := range []struct {
		sql  string
		dest any
	}{{"SELECT sqlite_version()", &d.SQLiteVersion}, {"PRAGMA journal_mode", &d.JournalMode}, {"PRAGMA foreign_keys", &d.ForeignKeys}, {"PRAGMA synchronous", &d.Synchronous}, {"SELECT max(version) FROM schema_migrations", &d.SchemaVersion}} {
		if s.db.QueryRowContext(ctx, p.sql).Scan(p.dest) != nil {
			return d, ErrStorage
		}
	}
	return d, nil
}
func NormalizeUsername(name string) (string, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if len(name) < 1 || len(name) > 64 {
		return "", ErrInvalid
	}
	for _, r := range name {
		if !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '.' || r == '_' || r == '-') {
			return "", ErrInvalid
		}
	}
	return name, nil
}

// Driver codes are normalized before crossing the Storage boundary.
func safeSQL(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return Error("STORAGE_CANCELLED")
	}
	var driver *sqliteDriver.Error
	if errors.As(err, &driver) {
		switch driver.Code() & 255 {
		case 5, 6:
			return Error("STORAGE_BUSY")
		case 8:
			return Error("STORAGE_READ_ONLY")
		case 11, 26:
			return Error("STORAGE_CORRUPT")
		case 13:
			return Error("STORAGE_FULL")
		case 19:
			return Error("STORAGE_CONSTRAINT")
		}
	}
	return ErrStorage
}
