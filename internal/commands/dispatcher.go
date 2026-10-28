// Package commands is the trusted internal admission boundary, not an HTTP gateway.
package commands

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/mPaprotski/Smart_home_hub/internal/storage"
	"strings"
	"time"
)

type Error string

func (e Error) Error() string { return string(e) }

const (
	ErrForbidden   Error = "FORBIDDEN"
	ErrInvalid     Error = "INVALID_COMMAND"
	ErrConflict    Error = "IDEMPOTENCY_CONFLICT"
	ErrExpired     Error = "EXPIRED"
	ErrUnsupported Error = "NOT_IMPLEMENTED"
	ErrOverride    Error = "MANUAL_OVERRIDE"
)

// Principal is supplied ONLY by a trusted gateway. External requests cannot set it.
type Principal struct {
	Credential *storage.Credential
	ID         string
	UserID     string
	Scopes     []string
	Bindings   []string
	Kind       string
}
type Command struct {
	ID               string
	Version          int
	Type             string
	Target           string
	Key              string
	Value            string
	ExpectedRevision int64
	ExpiresAt        time.Time
	CorrelationID    string
}
type Outcome struct {
	Status   string `json:"status"`
	Revision int64  `json:"revision"`
}
type Dispatcher struct {
	store *storage.SQLite
	now   func() time.Time
}

func New(store *storage.SQLite) *Dispatcher { return &Dispatcher{store: store, now: time.Now} }
func validSource(source string) bool {
	switch source {
	case "WEB_UI", "LOCAL_CLI", "BUTTON", "GESTURE", "TELEGRAM", "EXTERNAL_API", "AUTOMATION", "PREDICTION", "SYSTEM":
		return true
	}
	return false
}
func (d *Dispatcher) Execute(ctx context.Context, principal Principal, source string, c Command) (Outcome, error) {
	var result Outcome
	if c.Version != 1 || c.ID == "" || len(c.ID) > 128 || len(c.CorrelationID) > 128 || principal.UserID == "" || principal.ID == "" || !validSource(source) || c.Target == "" || len(c.Target) > 128 || c.ExpectedRevision < 0 || c.ExpiresAt.IsZero() {
		return result, ErrInvalid
	}
	if !d.now().Before(c.ExpiresAt) {
		return result, ErrExpired
	}
	// The first supported persistent command has no physical side effect.
	if c.Type != "SetUserSetting" {
		return result, ErrUnsupported
	}
	if storage.ValidateSetting(c.Key, c.Value) != nil {
		return result, ErrInvalid
	}
	if principal.Kind == "user" && principal.ID != principal.UserID {
		return result, ErrForbidden
	}
	if principal.Kind != "user" && principal.Kind != "service" {
		return result, ErrForbidden
	}
	if principal.UserID != c.Target {
		return result, ErrForbidden
	}
	for _, scope := range principal.Scopes {
		if scope != "settings:write" && scope != "settings:read" && scope != "lights:write" {
			return result, ErrForbidden
		}
	}
	if principal.Kind == "service" {
		scopeOK, bindingOK := false, false
		for _, scope := range principal.Scopes {
			if scope == "settings:write" {
				scopeOK = true
			} else if scope != "settings:read" && scope != "lights:write" {
				return result, ErrForbidden
			}
		}
		for _, binding := range principal.Bindings {
			if binding == "user:"+c.Target {
				bindingOK = true
			}
		}
		if !scopeOK || !bindingOK {
			return result, ErrForbidden
		}
	}
	normalized, _ := json.Marshal(struct {
		Type, Target, Key, Value string
		Revision                 int64
	}{c.Type, c.Target, c.Key, c.Value, c.ExpectedRevision})
	sum := sha256.Sum256(normalized)
	hash := hex.EncodeToString(sum[:])
	err := d.store.Transaction(ctx, func(tx *sql.Tx) error {
		if principal.Credential != nil {
			actual, err := storage.CheckCredentialTx(ctx, tx, *principal.Credential, d.now())
			if err != nil {
				return err
			}
			if actual.UserID != principal.UserID {
				return ErrForbidden
			}
			if actual.Kind == "token" {
				scopeOK, bindingOK := false, false
				for _, scope := range actual.Scopes {
					if scope == "settings:write" {
						scopeOK = true
					}
				}
				for _, binding := range actual.Bindings {
					if binding == "user:"+c.Target {
						bindingOK = true
					}
				}
				if actual.ID != principal.ID || !scopeOK || !bindingOK {
					return ErrForbidden
				}
			} else if principal.ID != actual.UserID {
				return ErrForbidden
			}
		}
		if err := storage.AuthorizeTx(ctx, tx, principal.UserID, "settings:write"); err != nil {
			return err
		}
		// Only resolved terminal metadata expires; no pending operations are removed.
		if _, err := tx.ExecContext(ctx, "DELETE FROM command_records WHERE status IN('APPLIED','FAILED','CANCELLED','EXPIRED') AND created_at<=?", d.now().Add(-24*time.Hour).Unix()); err != nil {
			return storage.ErrStorage
		}
		// Check live rights BEFORE reading a cached successful outcome.
		var previousHash, previousResult string
		err := tx.QueryRowContext(ctx, "SELECT request_hash,result FROM command_records WHERE principal_id=? AND operation=? AND command_id=?", principal.ID, c.Type, c.ID).Scan(&previousHash, &previousResult)
		if err == nil {
			if previousHash != hash {
				return ErrConflict
			}
			if json.Unmarshal([]byte(previousResult), &result) != nil {
				return storage.ErrStorage
			}
			return nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return storage.ErrStorage
		}
		setting, err := storage.SetSettingTx(ctx, tx, c.Target, c.Key, c.Value, c.ExpectedRevision)
		if err != nil {
			return err
		}
		if err := storage.AuditTx(ctx, tx, d.now(), "setting_update", "success", principal.UserID, "user_settings.timezone", "committed"); err != nil {
			return err
		}
		result = Outcome{Status: "APPLIED", Revision: setting.Revision}
		data, _ := json.Marshal(result)
		if _, err = tx.ExecContext(ctx, "INSERT INTO command_records VALUES(?,?,?,?,?,?,?)", principal.ID, c.ID, c.Type, hash, result.Status, string(data), d.now().Unix()); err != nil {
			return storage.ErrStorage
		}
		return nil
	})
	if err != nil {
		return Outcome{}, err
	}
	return result, nil
}

// Arbitration is the reusable policy port for future lighting handlers. Paths are
// server-derived; safety/capability/authorization must pass before calling it.
type Intent struct {
	PrincipalID, CommandID, Source string
	Paths                          []string
}

func AdmitLighting(ctx context.Context, tx *sql.Tx, intent Intent, now time.Time) error {
	if intent.PrincipalID == "" || intent.CommandID == "" || !validSource(intent.Source) || len(intent.Paths) == 0 || len(intent.Paths) > 32 {
		return ErrInvalid
	}
	manual := intent.Source != "AUTOMATION" && intent.Source != "PREDICTION" && intent.Source != "SYSTEM"
	for _, path := range intent.Paths {
		if path == "" || len(path) > 128 || !strings.HasPrefix(path, "lighting:") {
			return ErrInvalid
		}
		rows, err := tx.QueryContext(ctx, "SELECT path FROM manual_overrides WHERE expires_at>?", now.Unix())
		if err != nil {
			return storage.ErrStorage
		}
		blocked := false
		for rows.Next() {
			var held string
			if rows.Scan(&held) != nil {
				rows.Close()
				return storage.ErrStorage
			}
			if overlaps(held, path) {
				blocked = true
			}
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return storage.ErrStorage
		}
		if blocked && !manual {
			return ErrOverride
		}
	}
	if !manual {
		return nil
	}
	for _, path := range intent.Paths {
		if _, err := tx.ExecContext(ctx, "INSERT INTO manual_overrides VALUES(?,?,?,?) ON CONFLICT(path) DO UPDATE SET principal_id=excluded.principal_id,command_id=excluded.command_id,expires_at=excluded.expires_at WHERE manual_overrides.command_id<>excluded.command_id OR manual_overrides.principal_id<>excluded.principal_id", path, intent.PrincipalID, intent.CommandID, now.Add(900*time.Second).Unix()); err != nil {
			return storage.ErrStorage
		}
	}
	return nil
}
func overlaps(a, b string) bool {
	return a == b || strings.HasSuffix(a, ":*") && strings.HasPrefix(b, strings.TrimSuffix(a, "*")) || strings.HasSuffix(b, ":*") && strings.HasPrefix(a, strings.TrimSuffix(b, "*"))
}
