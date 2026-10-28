CREATE TABLE command_records(principal_id TEXT NOT NULL, command_id TEXT NOT NULL, operation TEXT NOT NULL, request_hash TEXT NOT NULL, status TEXT NOT NULL, result TEXT NOT NULL, created_at INTEGER NOT NULL, PRIMARY KEY(principal_id,operation,command_id));
CREATE INDEX command_retention ON command_records(status,created_at);
CREATE TABLE manual_overrides(path TEXT PRIMARY KEY, principal_id TEXT NOT NULL, command_id TEXT NOT NULL, expires_at INTEGER NOT NULL);
