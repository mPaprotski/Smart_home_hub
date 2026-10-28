ALTER TABLE sessions ADD COLUMN created_at INTEGER NOT NULL DEFAULT 0;
ALTER TABLE sessions ADD COLUMN last_seen_at INTEGER NOT NULL DEFAULT 0;
ALTER TABLE sessions ADD COLUMN csrf_state TEXT NOT NULL DEFAULT '';
CREATE INDEX sessions_active ON sessions(user_id,revoked_at,created_at);
