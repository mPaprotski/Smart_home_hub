// Purpose-specific domain repositories; none is a generic Event Bus journal.
package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"time"
)

func boundedName(s string) bool {
	if len(s) < 1 || len(s) > 64 {
		return false
	}
	for _, r := range s {
		if r < 32 {
			return false
		}
	}
	return true
}
func enabledUser(ctx context.Context, tx *sql.Tx, user string) error {
	var ok bool
	err := tx.QueryRowContext(ctx, "SELECT enabled FROM users WHERE id=?", user).Scan(&ok)
	if errors.Is(err, sql.ErrNoRows) {
		return Error("FORBIDDEN")
	}
	if err != nil {
		return ErrStorage
	}
	if !ok {
		return Error("FORBIDDEN")
	}
	return nil
}
func zoneAccess(ctx context.Context, tx *sql.Tx, user, zone string) error {
	var owner sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT owner_id FROM zones WHERE id=?", zone).Scan(&owner)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return ErrStorage
	}
	if owner.Valid && owner.String != user {
		return Error("FORBIDDEN")
	}
	return AuthorizeTx(ctx, tx, user, "lights:write")
}

type PresetEntry struct {
	ZoneID     string
	Mode       string
	Brightness float64
}

func (s *SQLite) CreatePreset(ctx context.Context, user, id, name string, entries []PresetEntry) error {
	if !boundedName(id) || !boundedName(name) || len(entries) < 1 || len(entries) > 64 {
		return ErrInvalid
	}
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		if err := enabledUser(ctx, tx, user); err != nil {
			return err
		}
		seen := map[string]bool{}
		for _, entry := range entries {
			if entry.Mode != "OFF" && entry.Mode != "STATIC" || !finiteRange(entry.Brightness, 0, 1) || seen[entry.ZoneID] {
				return ErrInvalid
			}
			seen[entry.ZoneID] = true
			if err := zoneAccess(ctx, tx, user, entry.ZoneID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, "INSERT INTO presets VALUES(?,?,?,1)", id, user, name); err != nil {
			return safeSQL(err)
		}
		for _, entry := range entries {
			if _, err := tx.ExecContext(ctx, "INSERT INTO preset_entries VALUES(?,?,?,?)", id, entry.ZoneID, entry.Mode, entry.Brightness); err != nil {
				return safeSQL(err)
			}
		}
		return nil
	})
}
func (s *SQLite) PresetEntries(ctx context.Context, user, id string) ([]PresetEntry, error) {
	out := []PresetEntry{}
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		var owner string
		err := tx.QueryRowContext(ctx, "SELECT owner_id FROM presets WHERE id=?", id).Scan(&owner)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return ErrStorage
		}
		if owner != user {
			return Error("FORBIDDEN")
		}
		rows, err := tx.QueryContext(ctx, "SELECT zone_id,mode,brightness FROM preset_entries WHERE preset_id=? ORDER BY zone_id LIMIT 64", id)
		if err != nil {
			return ErrStorage
		}
		defer rows.Close()
		for rows.Next() {
			var e PresetEntry
			if rows.Scan(&e.ZoneID, &e.Mode, &e.Brightness) != nil {
				return ErrStorage
			}
			out = append(out, e)
		}
		return rows.Err()
	})
	return out, err
}

type RuleDSL struct {
	Version int `json:"version"`
	Trigger struct {
		Kind     string  `json:"kind"`
		SensorID string  `json:"sensor_id"`
		Above    float64 `json:"above"`
	} `json:"trigger"`
	Action struct {
		Type       string  `json:"type"`
		ZoneID     string  `json:"zone_id"`
		Mode       string  `json:"mode"`
		Brightness float64 `json:"brightness"`
	} `json:"action"`
}

func ParseRule(data []byte) (RuleDSL, error) {
	var d RuleDSL
	if len(data) > 4096 {
		return d, ErrInvalid
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&d) != nil {
		return d, ErrInvalid
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return d, ErrInvalid
	}
	if d.Version != 1 || d.Trigger.Kind != "sensor_threshold" || !boundedName(d.Trigger.SensorID) || math.IsNaN(d.Trigger.Above) || math.IsInf(d.Trigger.Above, 0) || d.Action.Type != "SetZoneState" || !boundedName(d.Action.ZoneID) || d.Action.Mode != "OFF" && d.Action.Mode != "STATIC" || !finiteRange(d.Action.Brightness, 0, 1) {
		return d, ErrInvalid
	}
	return d, nil
}
func (s *SQLite) SaveRule(ctx context.Context, user, id string, data []byte, expected int64) error {
	d, err := ParseRule(data)
	if err != nil || !boundedName(id) || expected < 0 {
		return ErrInvalid
	}
	canonical, _ := json.Marshal(d)
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		if err := enabledUser(ctx, tx, user); err != nil {
			return err
		}
		if err := zoneAccess(ctx, tx, user, d.Action.ZoneID); err != nil {
			return err
		}
		var result sql.Result
		var err error
		if expected == 0 {
			result, err = tx.ExecContext(ctx, "INSERT INTO automation_rules VALUES(?,?,0,1,1,?) ON CONFLICT(id) DO NOTHING", id, user, string(canonical))
		} else {
			result, err = tx.ExecContext(ctx, "UPDATE automation_rules SET dsl=?,revision=revision+1 WHERE id=? AND owner_id=? AND revision=?", string(canonical), id, user, expected)
		}
		if err != nil {
			return safeSQL(err)
		}
		n, _ := result.RowsAffected()
		if n != 1 {
			return ErrConflict
		}
		return nil
	})
}

var gestures = map[string]bool{"OPEN_PALM": true, "SWIPE_RIGHT": true, "SWIPE_LEFT": true, "THUMB_UP": true, "THUMB_DOWN": true, "FIST": true, "TWO_FINGERS": true, "PALM_FACE_DOWN": true, "ROTATE_CLOCKWISE": true, "ROTATE_COUNTERCLOCKWISE": true}

func (s *SQLite) MapGesture(ctx context.Context, user, gesture, zone string, cooldown int, confirm bool) error {
	if !gestures[gesture] || cooldown < 0 || cooldown > 60000 {
		return ErrInvalid
	}
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		if err := zoneAccess(ctx, tx, user, zone); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO gesture_mappings VALUES(?,?,'SetZoneState',?,?,?) ON CONFLICT(user_id,gesture) DO UPDATE SET target=excluded.target,cooldown_ms=excluded.cooldown_ms,confirmation=excluded.confirmation", user, gesture, zone, cooldown, confirm)
		return safeSQL(err)
	})
}
func (s *SQLite) MapButton(ctx context.Context, user, device, press, zone string) error {
	if press != "single" && press != "double" && press != "long" {
		return ErrInvalid
	}
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		if err := zoneAccess(ctx, tx, user, zone); err != nil {
			return err
		}
		var owner string
		err := tx.QueryRowContext(ctx, "SELECT owner_id FROM button_mappings WHERE device_id=? AND press_kind=?", device, press).Scan(&owner)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return ErrStorage
		}
		if err == nil && owner != user {
			return Error("FORBIDDEN")
		}
		_, err = tx.ExecContext(ctx, "INSERT INTO button_mappings VALUES(?,?,?,'SetZoneState',?) ON CONFLICT(device_id,press_kind) DO UPDATE SET target=excluded.target", device, press, user, zone)
		return safeSQL(err)
	})
}
func finiteRange(v, low, high float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= low && v <= high
}

type HabitObservation struct {
	ID                 int64
	At                 time.Time
	Timezone           string
	Epoch              int64
	ZoneID, Mode       string
	ColorRGB           *int
	Brightness, Volume *float64
	Presence           *string
	AmbientLux         *float64
	MonitorActive      *bool
	Origin, Quality    string
}

func (h HabitObservation) Validate() error {
	if h.At.IsZero() || h.Epoch < 1 || !boundedName(h.ZoneID) || h.Mode != "OFF" && h.Mode != "STATIC" || h.Origin != "manual" && h.Origin != "automation" && h.Origin != "prediction" || h.Quality != "valid" && h.Quality != "stale" && h.Quality != "unknown" {
		return ErrInvalid
	}
	if _, err := time.LoadLocation(h.Timezone); err != nil || len(h.Timezone) > 64 {
		return ErrInvalid
	}
	if h.ColorRGB != nil && (*h.ColorRGB < 0 || *h.ColorRGB > 16777215) {
		return ErrInvalid
	}
	for _, v := range []*float64{h.Brightness, h.Volume} {
		if v != nil && !finiteRange(*v, 0, 1) {
			return ErrInvalid
		}
	}
	if h.AmbientLux != nil && !finiteRange(*h.AmbientLux, 0, 1e9) {
		return ErrInvalid
	}
	if h.Presence != nil && *h.Presence != "PRESENT" && *h.Presence != "PROBABLY_PRESENT" && *h.Presence != "AWAY" && *h.Presence != "UNKNOWN" {
		return ErrInvalid
	}
	return nil
}
func (s *SQLite) StartLearning(ctx context.Context, user string, now time.Time) error {
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		if err := enabledUser(ctx, tx, user); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO habit_learning_state VALUES(?,?,1,'insufficient') ON CONFLICT(user_id) DO NOTHING", user, now.Unix())
		return safeSQL(err)
	})
}
func currentEpoch(ctx context.Context, tx *sql.Tx, user string, want int64) error {
	var epoch int64
	if err := enabledUser(ctx, tx, user); err != nil {
		return err
	}
	err := tx.QueryRowContext(ctx, "SELECT privacy_epoch FROM habit_learning_state WHERE user_id=?", user).Scan(&epoch)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return ErrStorage
	}
	if epoch != want {
		return ErrConflict
	}
	return nil
}
func (s *SQLite) AddHabit(ctx context.Context, user string, h HabitObservation) error {
	if err := h.Validate(); err != nil {
		return err
	}
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		if err := currentEpoch(ctx, tx, user, h.Epoch); err != nil {
			return err
		}
		if err := zoneAccess(ctx, tx, user, h.ZoneID); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO habit_observations(user_id,occurred_at,timezone,privacy_epoch,zone_id,mode,color_rgb,brightness,volume,presence,ambient_lux,monitor_active,origin,quality) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)", user, h.At.UTC().Unix(), h.Timezone, h.Epoch, h.ZoneID, h.Mode, h.ColorRGB, h.Brightness, h.Volume, h.Presence, h.AmbientLux, h.MonitorActive, h.Origin, h.Quality)
		return safeSQL(err)
	})
}
func (s *SQLite) Habits(ctx context.Context, user string, afterID int64, limit int, now time.Time) ([]HabitObservation, error) {
	if limit < 1 || limit > 100 || afterID < 0 {
		return nil, ErrInvalid
	}
	out := []HabitObservation{}
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		if err := enabledUser(ctx, tx, user); err != nil {
			return err
		}
		rows, err := tx.QueryContext(ctx, "SELECT id,occurred_at,timezone,privacy_epoch,zone_id,mode,color_rgb,brightness,volume,presence,ambient_lux,monitor_active,origin,quality FROM habit_observations WHERE user_id=? AND id>? AND occurred_at>=? AND privacy_epoch=(SELECT privacy_epoch FROM habit_learning_state WHERE user_id=?) ORDER BY id LIMIT ?", user, afterID, now.UTC().AddDate(0, 0, -90).Unix(), user, limit)
		if err != nil {
			return ErrStorage
		}
		defer rows.Close()
		for rows.Next() {
			var h HabitObservation
			var at int64
			if rows.Scan(&h.ID, &at, &h.Timezone, &h.Epoch, &h.ZoneID, &h.Mode, &h.ColorRGB, &h.Brightness, &h.Volume, &h.Presence, &h.AmbientLux, &h.MonitorActive, &h.Origin, &h.Quality) != nil {
				return ErrStorage
			}
			h.At = time.Unix(at, 0).UTC()
			out = append(out, h)
		}
		return rows.Err()
	})
	return out, err
}
func (s *SQLite) PublishModel(ctx context.Context, user, version, checksum string, epoch int64, at time.Time) error {
	if !boundedName(version) || len(checksum) != 64 {
		return ErrInvalid
	}
	if _, err := hex.DecodeString(checksum); err != nil {
		return ErrInvalid
	}
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		if err := currentEpoch(ctx, tx, user, epoch); err != nil {
			return err
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO habit_model_metadata VALUES(?,?,?,?,?)", user, version, epoch, checksum, at.Unix())
		return safeSQL(err)
	})
}

type ProposedAction struct {
	ZoneID, Mode       string
	ColorRGB           *int
	Brightness, Volume *float64
}

func (s *SQLite) Recommend(ctx context.Context, user, id, model string, epoch int64, confidence float64, action ProposedAction, expires, now time.Time) error {
	if !boundedName(id) || !finiteRange(confidence, 0, 1) || !expires.After(now) || expires.After(now.Add(24*time.Hour)) {
		return ErrInvalid
	}
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		if err := currentEpoch(ctx, tx, user, epoch); err != nil {
			return err
		}
		if err := zoneAccess(ctx, tx, user, action.ZoneID); err != nil {
			return err
		}
		if (HabitObservation{At: now, Timezone: "UTC", Epoch: epoch, ZoneID: action.ZoneID, Mode: action.Mode, ColorRGB: action.ColorRGB, Brightness: action.Brightness, Volume: action.Volume, Origin: "prediction", Quality: "valid"}).Validate() != nil {
			return ErrInvalid
		}
		var modelEpoch int64
		if tx.QueryRowContext(ctx, "SELECT privacy_epoch FROM habit_model_metadata WHERE user_id=? AND version=?", user, model).Scan(&modelEpoch) != nil {
			return ErrNotFound
		}
		if modelEpoch != epoch {
			return ErrConflict
		}
		_, err := tx.ExecContext(ctx, "INSERT INTO recommendations VALUES(?,?,?,?,?,'SetZoneState',?,?,?,?,?,?,'pending',?)", id, user, epoch, model, confidence, action.ZoneID, action.Mode, action.ColorRGB, action.Brightness, action.Volume, now.Unix(), expires.Unix())
		return safeSQL(err)
	})
}
func (s *SQLite) Feedback(ctx context.Context, user, id, decision string, now time.Time) error {
	if decision != "accepted" && decision != "rejected" {
		return ErrInvalid
	}
	return s.Transaction(ctx, func(tx *sql.Tx) error {
		var owner, status string
		var epoch, expires int64
		err := tx.QueryRowContext(ctx, "SELECT user_id,privacy_epoch,status,expires_at FROM recommendations WHERE id=?", id).Scan(&owner, &epoch, &status, &expires)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return ErrStorage
		}
		if owner != user {
			return Error("FORBIDDEN")
		}
		if err = currentEpoch(ctx, tx, user, epoch); err != nil {
			return err
		}
		var previous string
		err = tx.QueryRowContext(ctx, "SELECT decision FROM prediction_feedback WHERE user_id=? AND recommendation_id=?", user, id).Scan(&previous)
		if err == nil {
			if previous == decision {
				return nil
			}
			return ErrConflict
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return ErrStorage
		}
		if expires <= now.Unix() || status != "pending" {
			return ErrConflict
		}
		if _, err = tx.ExecContext(ctx, "INSERT INTO prediction_feedback VALUES(?,?,?,?,?)", id, user, epoch, decision, now.Unix()); err != nil {
			return safeSQL(err)
		}
		_, err = tx.ExecContext(ctx, "UPDATE recommendations SET status=? WHERE id=?", decision, id)
		return safeSQL(err)
	})
}

// Cleanup executes bounded batches; callers repeat later, not a single huge delete.
func (s *SQLite) Cleanup(ctx context.Context, now time.Time, batch int) (int64, error) {
	if batch < 1 || batch > 1000 {
		return 0, ErrInvalid
	}
	var removed int64
	err := s.Transaction(ctx, func(tx *sql.Tx) error {
		for _, table := range []struct {
			name, clock string
			cutoff      int64
		}{{"habit_observations", "occurred_at", now.UTC().AddDate(0, 0, -90).Unix()}, {"automation_audit", "occurred_at", now.UTC().AddDate(0, 0, -30).Unix()}, {"security_audit", "occurred_at", now.UTC().AddDate(0, 0, -30).Unix()}} {
			result, err := tx.ExecContext(ctx, "DELETE FROM "+table.name+" WHERE id IN(SELECT id FROM "+table.name+" WHERE "+table.clock+"<? ORDER BY "+table.clock+" LIMIT ?)", table.cutoff, batch)
			if err != nil {
				return safeSQL(err)
			}
			n, _ := result.RowsAffected()
			removed += n
		}
		if _, err := tx.ExecContext(ctx, "UPDATE habit_learning_state SET quality='insufficient' WHERE user_id IN(SELECT user_id FROM habit_model_metadata WHERE trained_at<? LIMIT ?)", now.UTC().AddDate(0, 0, -90).Unix(), batch); err != nil {
			return safeSQL(err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM recommendations WHERE id IN(SELECT id FROM recommendations WHERE created_at<? OR (user_id,model_version) IN(SELECT user_id,version FROM habit_model_metadata WHERE trained_at<?) LIMIT ?)", now.UTC().AddDate(0, 0, -90).Unix(), now.UTC().AddDate(0, 0, -90).Unix(), batch); err != nil {
			return safeSQL(err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM habit_model_metadata WHERE rowid IN(SELECT rowid FROM habit_model_metadata m WHERE trained_at<? AND NOT EXISTS(SELECT 1 FROM recommendations r WHERE r.user_id=m.user_id AND r.model_version=m.version) LIMIT ?)", now.UTC().AddDate(0, 0, -90).Unix(), batch); err != nil {
			return safeSQL(err)
		}
		if _, err := tx.ExecContext(ctx, "UPDATE recommendations SET status='expired' WHERE id IN(SELECT id FROM recommendations WHERE status='pending' AND expires_at<=? LIMIT ?)", now.Unix(), batch); err != nil {
			return safeSQL(err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM sessions WHERE id_hash IN(SELECT id_hash FROM sessions WHERE expires_at<=? OR revoked_at IS NOT NULL OR last_seen_at<=? LIMIT ?)", now.Unix(), now.Add(-SessionIdle).Unix(), batch); err != nil {
			return safeSQL(err)
		}
		if _, err := tx.ExecContext(ctx, "DELETE FROM api_tokens WHERE id IN(SELECT id FROM api_tokens WHERE expires_at<=? OR revoked_at IS NOT NULL LIMIT ?)", now.Unix(), batch); err != nil {
			return safeSQL(err)
		}
		return nil
	})
	return removed, err
}
func (s *SQLite) ObservationQueryPlan(ctx context.Context, user string, after time.Time) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, "EXPLAIN QUERY PLAN SELECT id FROM habit_observations WHERE user_id=? AND occurred_at>=? ORDER BY occurred_at LIMIT 100", user, after.Unix())
	if err != nil {
		return nil, ErrStorage
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if rows.Scan(&id, &parent, &unused, &detail) != nil {
			return nil, ErrStorage
		}
		out = append(out, detail)
	}
	return out, rows.Err()
}
