package storage

import (
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

const SessionIdle = 30 * time.Minute
const SessionAbsolute = 12 * time.Hour

// Credential is an internal proof/reference. Never JSON-encode it or include it in logs.
type Credential struct {
	Kind, ID, UserID, Proof, CSRF    string
	Scopes, Bindings                 []string
	CreatedAt, LastSeenAt, ExpiresAt int64
}

func AuditTx(ctx context.Context, tx *sql.Tx, at time.Time, action, result, principal, resource, reason string) error {
	allowed := map[string]bool{"login": true, "login_failed": true, "logout": true, "sessions_revoke": true, "token_create": true, "token_revoke": true, "token_denied": true, "setting_update": true}
	if !allowed[action] || (result != "success" && result != "denied") || len(principal) > 128 || len(resource) > 128 || len(reason) > 64 {
		return ErrInvalid
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO security_audit(occurred_at,action,result,principal_id,resource_ref,reason) VALUES(?,?,?,?,?,?)", at.Unix(), action, result, principal, resource, reason); err != nil {
		return ErrStorage
	}
	return nil
}
func CheckCredentialTx(ctx context.Context, tx *sql.Tx, c Credential, now time.Time) (Credential, error) {
	var actual Credential
	var enabled bool
	var revoked sql.NullInt64
	if c.Kind == "session" {
		actual.Kind = "session"
		actual.ID = c.ID
		err := tx.QueryRowContext(ctx, "SELECT s.user_id,s.created_at,s.last_seen_at,s.expires_at,s.revoked_at,s.csrf_state,u.enabled FROM sessions s JOIN users u ON u.id=s.user_id WHERE s.id_hash=?", c.ID).Scan(&actual.UserID, &actual.CreatedAt, &actual.LastSeenAt, &actual.ExpiresAt, &revoked, &actual.CSRF, &enabled)
		if errors.Is(err, sql.ErrNoRows) {
			return actual, Error("UNAUTHENTICATED")
		}
		if err != nil {
			return actual, ErrStorage
		}
		if !enabled || revoked.Valid || now.Unix() >= actual.ExpiresAt || now.Unix() >= actual.LastSeenAt+int64(SessionIdle/time.Second) || actual.CSRF == "" {
			return Credential{}, Error("UNAUTHENTICATED")
		}
	} else if c.Kind == "token" {
		var proof, scopes, bindings string
		actual.Kind = "token"
		actual.ID = c.ID
		err := tx.QueryRowContext(ctx, "SELECT t.owner_id,t.secret_hash,t.scopes,t.bindings,t.created_at,t.expires_at,t.revoked_at,u.enabled FROM api_tokens t JOIN users u ON u.id=t.owner_id WHERE t.id=?", c.ID).Scan(&actual.UserID, &proof, &scopes, &bindings, &actual.CreatedAt, &actual.ExpiresAt, &revoked, &enabled)
		if errors.Is(err, sql.ErrNoRows) {
			return actual, Error("UNAUTHENTICATED")
		}
		if err != nil {
			return actual, ErrStorage
		}
		if !enabled || revoked.Valid || now.Unix() >= actual.ExpiresAt || subtle.ConstantTimeCompare([]byte(proof), []byte(c.Proof)) != 1 {
			return Credential{}, Error("UNAUTHENTICATED")
		}
		if json.Unmarshal([]byte(scopes), &actual.Scopes) != nil || json.Unmarshal([]byte(bindings), &actual.Bindings) != nil || len(actual.Scopes) > 8 || len(actual.Bindings) > 16 {
			return actual, ErrStorage
		}
		actual.Proof = c.Proof
	} else {
		return actual, Error("UNAUTHENTICATED")
	}
	if c.UserID != "" && c.UserID != actual.UserID {
		return Credential{}, Error("UNAUTHENTICATED")
	}
	return actual, nil
}
func (s *SQLite) IssueSession(ctx context.Context, user, digest, csrf string, now time.Time) error {
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		var enabled bool
		if tx.QueryRowContext(ctx, "SELECT enabled FROM users WHERE id=?", user).Scan(&enabled) != nil {
			return ErrStorage
		}
		if !enabled {
			return Error("UNAUTHENTICATED")
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE expires_at<=? OR revoked_at IS NOT NULL OR last_seen_at<=?", now.Unix(), now.Add(-SessionIdle).Unix()); err != nil {
			return ErrStorage
		}
		if _, err := tx.ExecContext(ctx, "UPDATE sessions SET revoked_at=? WHERE user_id=? AND revoked_at IS NULL AND id_hash NOT IN(SELECT id_hash FROM sessions WHERE user_id=? AND revoked_at IS NULL ORDER BY created_at DESC,id_hash LIMIT 4)", now.Unix(), user, user); err != nil {
			return ErrStorage
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO sessions(id_hash,user_id,expires_at,created_at,last_seen_at,csrf_state) VALUES(?,?,?,?,?,?)", digest, user, now.Add(SessionAbsolute).Unix(), now.Unix(), now.Unix(), csrf); err != nil {
			return ErrStorage
		}
		return AuditTx(ctx, tx, now, "login", "success", user, "session", "created")
	})
}
func (s *SQLite) CheckCredential(ctx context.Context, c Credential, now time.Time) (Credential, error) {
	var actual Credential
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		var err error
		actual, err = CheckCredentialTx(ctx, tx, c, now)
		if err != nil {
			return err
		}
		if actual.Kind == "session" {
			if _, err := tx.ExecContext(ctx, "UPDATE sessions SET last_seen_at=? WHERE id_hash=?", now.Unix(), actual.ID); err != nil {
				return ErrStorage
			}
			actual.LastSeenAt = now.Unix()
		}
		return nil
	})
	return actual, err
}
func (s *SQLite) RevokeSessions(ctx context.Context, c Credential, all bool, now time.Time) error {
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		actual, err := CheckCredentialTx(ctx, tx, c, now)
		if err != nil {
			return err
		}
		if actual.Kind != "session" {
			return Error("FORBIDDEN")
		}
		query := "UPDATE sessions SET revoked_at=? WHERE id_hash=?"
		key := actual.ID
		action := "logout"
		if all {
			query = "UPDATE sessions SET revoked_at=? WHERE user_id=?"
			key = actual.UserID
			action = "sessions_revoke"
		}
		if _, err := tx.ExecContext(ctx, query, now.Unix(), key); err != nil {
			return ErrStorage
		}
		return AuditTx(ctx, tx, now, action, "success", actual.UserID, "sessions", "revoked")
	})
}

type TokenMetadata struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	Bindings  []string `json:"bindings"`
	CreatedAt int64    `json:"created_at"`
	ExpiresAt int64    `json:"expires_at"`
	RevokedAt *int64   `json:"revoked_at,omitempty"`
}

func (s *SQLite) CreateToken(ctx context.Context, c Credential, meta TokenMetadata, digest string, now time.Time) error {
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		actual, err := CheckCredentialTx(ctx, tx, c, now)
		if err != nil {
			return err
		}
		if actual.Kind != "session" {
			return Error("FORBIDDEN")
		}
		if err = AuthorizeTx(ctx, tx, actual.UserID, "tokens:manage"); err != nil {
			return err
		}
		for _, scope := range meta.Scopes {
			if scope != "settings:write" && scope != "settings:read" {
				return Error("FORBIDDEN")
			}
			if err = AuthorizeTx(ctx, tx, actual.UserID, scope); err != nil {
				return err
			}
		}
		for _, binding := range meta.Bindings {
			if binding != "user:"+actual.UserID {
				return Error("FORBIDDEN")
			}
		}
		var count int
		if tx.QueryRowContext(ctx, "SELECT count(*) FROM api_tokens WHERE owner_id=? AND revoked_at IS NULL AND expires_at>?", actual.UserID, now.Unix()).Scan(&count) != nil {
			return ErrStorage
		}
		if count >= 16 {
			return Error("TOKEN_LIMIT")
		}
		scopes, _ := json.Marshal(meta.Scopes)
		bindings, _ := json.Marshal(meta.Bindings)
		if _, err := tx.ExecContext(ctx, "INSERT INTO api_tokens(id,owner_id,secret_hash,name,scopes,bindings,created_at,expires_at) VALUES(?,?,?,?,?,?,?,?)", meta.ID, actual.UserID, digest, meta.Name, string(scopes), string(bindings), now.Unix(), meta.ExpiresAt); err != nil {
			return ErrStorage
		}
		return AuditTx(ctx, tx, now, "token_create", "success", actual.UserID, meta.ID, "created")
	})
}
func (s *SQLite) ListTokens(ctx context.Context, c Credential, now time.Time) ([]TokenMetadata, error) {
	out := []TokenMetadata{}
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		actual, err := CheckCredentialTx(ctx, tx, c, now)
		if err != nil {
			return err
		}
		if actual.Kind != "session" {
			return Error("FORBIDDEN")
		}
		if err = AuthorizeTx(ctx, tx, actual.UserID, "tokens:manage"); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, "SELECT id,name,scopes,bindings,created_at,expires_at,revoked_at FROM api_tokens WHERE owner_id=? ORDER BY created_at DESC,id LIMIT 100", actual.UserID)
		if err != nil {
			return ErrStorage
		}
		defer rows.Close()
		for rows.Next() {
			var t TokenMetadata
			var a, b string
			var revoked sql.NullInt64
			if rows.Scan(&t.ID, &t.Name, &a, &b, &t.CreatedAt, &t.ExpiresAt, &revoked) != nil {
				return ErrStorage
			}
			if json.Unmarshal([]byte(a), &t.Scopes) != nil || json.Unmarshal([]byte(b), &t.Bindings) != nil {
				return ErrStorage
			}
			if revoked.Valid {
				v := revoked.Int64
				t.RevokedAt = &v
			}
			out = append(out, t)
		}
		if rows.Err() != nil {
			return ErrStorage
		}
		return nil
	})
	return out, err
}
func (s *SQLite) RevokeToken(ctx context.Context, c Credential, id string, now time.Time) error {
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		actual, err := CheckCredentialTx(ctx, tx, c, now)
		if err != nil {
			return err
		}
		if actual.Kind != "session" {
			return Error("FORBIDDEN")
		}
		if err = AuthorizeTx(ctx, tx, actual.UserID, "tokens:manage"); err != nil {
			return err
		}
		var owner string
		err = tx.QueryRowContext(ctx, "SELECT owner_id FROM api_tokens WHERE id=?", id).Scan(&owner)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return ErrStorage
		}
		if owner != actual.UserID {
			return Error("FORBIDDEN")
		}
		if _, err = tx.ExecContext(ctx, "UPDATE api_tokens SET revoked_at=COALESCE(revoked_at,?) WHERE id=?", now.Unix(), id); err != nil {
			return ErrStorage
		}
		return AuditTx(ctx, tx, now, "token_revoke", "success", actual.UserID, id, "revoked")
	})
}

type AuditEntry struct {
	ID        int64  `json:"id"`
	At        int64  `json:"occurred_at"`
	Action    string `json:"action"`
	Result    string `json:"result"`
	Principal string `json:"principal"`
	Resource  string `json:"resource"`
	Reason    string `json:"reason"`
}

func (s *SQLite) ListAudit(ctx context.Context, c Credential, after int64, now time.Time) ([]AuditEntry, error) {
	out := []AuditEntry{}
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		actual, err := CheckCredentialTx(ctx, tx, c, now)
		if err != nil {
			return err
		}
		if actual.Kind != "session" {
			return Error("FORBIDDEN")
		}
		if err = AuthorizeTx(ctx, tx, actual.UserID, "audit:read"); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, "SELECT id,occurred_at,action,result,principal_id,resource_ref,reason FROM security_audit WHERE principal_id=? AND id>? ORDER BY id LIMIT 100", actual.UserID, after)
		if err != nil {
			return ErrStorage
		}
		defer rows.Close()
		for rows.Next() {
			var a AuditEntry
			if rows.Scan(&a.ID, &a.At, &a.Action, &a.Result, &a.Principal, &a.Resource, &a.Reason) != nil {
				return ErrStorage
			}
			out = append(out, a)
		}
		return rows.Err()
	})
	return out, err
}

// Repeated failed logins are aggregated to a bounded per-minute record, not raw inputs.
func (s *SQLite) AuditFailedLogin(ctx context.Context, now time.Time) error {
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		var n int
		if tx.QueryRowContext(ctx, "SELECT count(*) FROM security_audit WHERE action='login_failed' AND occurred_at>=?", now.Unix()/60*60).Scan(&n) != nil {
			return ErrStorage
		}
		if n > 0 {
			return nil
		}
		return AuditTx(ctx, tx, now, "login_failed", "denied", "", "auth", "invalid_credentials")
	})
}

func (s *SQLite) PeekCredential(ctx context.Context, c Credential, now time.Time) (Credential, error) {
	var actual Credential
	err := s.Transaction(ctx, func(tx *sql.Tx) error { var err error; actual, err = CheckCredentialTx(ctx, tx, c, now); return err })
	return actual, err
}
func (c Credential) String() string               { return "[REDACTED_CREDENTIAL]" }
func (c Credential) GoString() string             { return c.String() }
func (c Credential) MarshalJSON() ([]byte, error) { return []byte(`"[REDACTED_CREDENTIAL]"`), nil }

func (s *SQLite) AuditDenied(ctx context.Context, now time.Time, action string) error {
	if action != "login_failed" && action != "token_denied" && action != "setting_update" {
		return ErrInvalid
	}
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		var n int
		if tx.QueryRowContext(ctx, "SELECT count(*) FROM security_audit WHERE action=? AND result='denied' AND occurred_at>=?", action, now.Unix()/60*60).Scan(&n) != nil {
			return ErrStorage
		}
		if n > 0 {
			return nil
		}
		return AuditTx(ctx, tx, now, action, "denied", "", "auth", "invalid_credentials")
	})
}
