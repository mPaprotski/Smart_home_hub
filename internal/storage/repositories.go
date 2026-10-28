package storage

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"math"
	"time"
)

type User struct {
	ID, Username string
	PasswordHash string `json:"-"`
	Enabled      bool
	Timezone     string
}

func (s *SQLite) FindUser(ctx context.Context, username string) (User, error) {
	var u User
	err := s.db.QueryRowContext(ctx, "SELECT id,username,password_hash,enabled,timezone FROM users WHERE username=?", username).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Enabled, &u.Timezone)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	if err != nil {
		return u, ErrStorage
	}
	return u, nil
}
func (s *SQLite) Bootstrap(ctx context.Context, username, hash string) (string, error) {
	name, err := NormalizeUsername(username)
	if err != nil || hash == "" || len(hash) > 512 {
		return "", ErrInvalid
	}
	var b [16]byte
	if _, err = rand.Read(b[:]); err != nil {
		return "", ErrStorage
	}
	id := hex.EncodeToString(b[:])
	err = s.Transaction(ctx, func(tx *sql.Tx) error {
		var count int
		if tx.QueryRowContext(ctx, "SELECT count(*) FROM bootstrap_state").Scan(&count) != nil {
			return ErrStorage
		}
		if count != 0 {
			return ErrInitialized
		}
		if tx.QueryRowContext(ctx, "SELECT count(*) FROM users").Scan(&count) != nil {
			return ErrStorage
		}
		if count != 0 {
			return ErrInitialized
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO users VALUES(?,?,?,1,'UTC',?)", id, name, hash, time.Now().Unix()); err != nil {
			return ErrStorage
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO user_roles VALUES(?,'admin')", id); err != nil {
			return ErrStorage
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO bootstrap_state VALUES(1,?)", time.Now().Unix()); err != nil {
			return ErrStorage
		}
		return nil
	})
	return id, err
}

type Setting struct {
	Value    string `json:"value"`
	Revision int64  `json:"revision"`
}

func ValidateSetting(key, value string) error {
	if key != "timezone" || value == "" || len(value) > 64 {
		return ErrInvalid
	}
	if _, err := time.LoadLocation(value); err != nil {
		return ErrInvalid
	}
	return nil
}
func GetSettingTx(ctx context.Context, tx *sql.Tx, user, key string) (Setting, error) {
	var s Setting
	err := tx.QueryRowContext(ctx, "SELECT value,revision FROM user_settings WHERE user_id=? AND key=?", user, key).Scan(&s.Value, &s.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return s, ErrNotFound
	}
	if err != nil {
		return s, ErrStorage
	}
	return s, nil
}
func SetSettingTx(ctx context.Context, tx *sql.Tx, user, key, value string, expected int64) (Setting, error) {
	if ValidateSetting(key, value) != nil || expected < 0 {
		return Setting{}, ErrInvalid
	}
	var result sql.Result
	var err error
	if expected == 0 {
		result, err = tx.ExecContext(ctx, "INSERT INTO user_settings VALUES(?,?,?,1,1) ON CONFLICT(user_id,key) DO NOTHING", user, key, value)
	} else {
		result, err = tx.ExecContext(ctx, "UPDATE user_settings SET value=?,revision=revision+1 WHERE user_id=? AND key=? AND revision=?", value, user, key, expected)
	}
	if err != nil {
		return Setting{}, ErrStorage
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Setting{}, ErrStorage
	}
	if n != 1 {
		return Setting{}, ErrConflict
	}
	return Setting{value, expected + 1}, nil
}
func (s *SQLite) GetSetting(ctx context.Context, user, key string) (Setting, error) {
	var result Setting
	err := s.Transaction(ctx, func(tx *sql.Tx) error { var err error; result, err = GetSettingTx(ctx, tx, user, key); return err })
	return result, err
}

// Current permissions are loaded inside the admission transaction, including retry.
func AuthorizeTx(ctx context.Context, tx *sql.Tx, user, scope string) error {
	if scope != "settings:write" && scope != "lights:write" && scope != "settings:read" && scope != "tokens:manage" && scope != "audit:read" {
		return Error("FORBIDDEN")
	}
	var n int
	if tx.QueryRowContext(ctx, "SELECT count(*) FROM users u JOIN user_roles ur ON ur.user_id=u.id JOIN role_permissions rp ON rp.role_id=ur.role_id WHERE u.id=? AND u.enabled=1 AND rp.scope=?", user, scope).Scan(&n) != nil {
		return ErrStorage
	}
	if n == 0 {
		return Error("FORBIDDEN")
	}
	return nil
}

// Persistent account cooldown survives restarts; keys are keyed digests, not usernames.
func (s *SQLite) Attempt(ctx context.Context, key string, now time.Time, success bool) (bool, error) {
	allowed := false
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		cutoff := now.Add(-15 * time.Minute).Unix()
		if _, err := tx.ExecContext(ctx, "DELETE FROM auth_attempts WHERE window_start<=?", cutoff); err != nil {
			return ErrStorage
		}
		if success {
			_, err := tx.ExecContext(ctx, "DELETE FROM auth_attempts WHERE account_key=?", key)
			if err != nil {
				return ErrStorage
			}
			allowed = true
			return nil
		}
		var failures int
		var start int64
		err := tx.QueryRowContext(ctx, "SELECT failures,window_start FROM auth_attempts WHERE account_key=?", key).Scan(&failures, &start)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return ErrStorage
		}
		if failures >= 5 {
			return nil
		}
		var count int
		if tx.QueryRowContext(ctx, "SELECT count(*) FROM auth_attempts").Scan(&count) != nil {
			return ErrStorage
		}
		if errors.Is(err, sql.ErrNoRows) && count >= 4096 {
			return nil
		}
		if failures == 0 {
			start = now.Unix()
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO auth_attempts VALUES(?,1,?) ON CONFLICT(account_key) DO UPDATE SET failures=failures+1", key, start); err != nil {
			return ErrStorage
		}
		allowed = true
		return nil
	})
	return allowed, err
}
func (s *SQLite) EnsureLimiterKey(ctx context.Context) ([]byte, error) {
	var key []byte
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		err := tx.QueryRowContext(ctx, "SELECT limiter_key FROM auth_internal WHERE singleton=1").Scan(&key)
		if err == nil {
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return ErrStorage
		}
		key = make([]byte, 32)
		if _, err = rand.Read(key); err != nil {
			return ErrStorage
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO auth_internal VALUES(1,?)", key); err != nil {
			return ErrStorage
		}
		return nil
	})
	return key, err
}

type Zone struct {
	ID, DeviceID, OwnerID, Mode string
	Brightness                  float64
	Revision                    int64
}

func SetZoneTx(ctx context.Context, tx *sql.Tx, z Zone, expected int64) (Zone, error) {
	if z.ID == "" || len(z.ID) > 128 || z.Mode != "OFF" && z.Mode != "STATIC" || math.IsNaN(z.Brightness) || math.IsInf(z.Brightness, 0) || z.Brightness < 0 || z.Brightness > 1 || expected < 0 {
		return Zone{}, ErrInvalid
	}
	nullable := func(s string) any {
		if s == "" {
			return nil
		}
		return s
	}
	var result sql.Result
	var err error
	if expected == 0 {
		result, err = tx.ExecContext(ctx, "INSERT INTO zones VALUES(?,?,?,?,?,1) ON CONFLICT(id) DO NOTHING", z.ID, nullable(z.DeviceID), nullable(z.OwnerID), z.Mode, z.Brightness)
	} else {
		result, err = tx.ExecContext(ctx, "UPDATE zones SET mode=?,brightness=?,revision=revision+1 WHERE id=? AND revision=?", z.Mode, z.Brightness, z.ID, expected)
	}
	if err != nil {
		return Zone{}, ErrStorage
	}
	n, err := result.RowsAffected()
	if err != nil {
		return Zone{}, ErrStorage
	}
	if n != 1 {
		return Zone{}, ErrConflict
	}
	var actual Zone
	var device, owner sql.NullString
	if tx.QueryRowContext(ctx, "SELECT id,device_id,owner_id,mode,brightness,revision FROM zones WHERE id=?", z.ID).Scan(&actual.ID, &device, &owner, &actual.Mode, &actual.Brightness, &actual.Revision) != nil {
		return Zone{}, ErrStorage
	}
	actual.DeviceID = device.String
	actual.OwnerID = owner.String
	return actual, nil
}
func (s *SQLite) GetZone(ctx context.Context, id string) (Zone, error) {
	var z Zone
	var device, owner sql.NullString
	err := s.db.QueryRowContext(ctx, "SELECT id,device_id,owner_id,mode,brightness,revision FROM zones WHERE id=?", id).Scan(&z.ID, &device, &owner, &z.Mode, &z.Brightness, &z.Revision)
	if errors.Is(err, sql.ErrNoRows) {
		return z, ErrNotFound
	}
	if err != nil {
		return z, ErrStorage
	}
	z.DeviceID = device.String
	z.OwnerID = owner.String
	return z, nil
}
