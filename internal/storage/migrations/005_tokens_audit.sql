CREATE TABLE api_tokens(id TEXT PRIMARY KEY, owner_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE, secret_hash TEXT NOT NULL UNIQUE, name TEXT NOT NULL CHECK(length(name)<=64), scopes TEXT NOT NULL, bindings TEXT NOT NULL, created_at INTEGER NOT NULL, expires_at INTEGER NOT NULL, revoked_at INTEGER);
CREATE INDEX token_owner ON api_tokens(owner_id,revoked_at,created_at);
CREATE INDEX token_expiry ON api_tokens(expires_at);
CREATE TABLE security_audit(id INTEGER PRIMARY KEY, occurred_at INTEGER NOT NULL, action TEXT NOT NULL, result TEXT NOT NULL, principal_id TEXT NOT NULL, resource_ref TEXT NOT NULL, reason TEXT NOT NULL);
CREATE INDEX security_audit_time ON security_audit(occurred_at);
CREATE INDEX security_audit_principal ON security_audit(principal_id,id);
INSERT INTO role_permissions VALUES('admin','tokens:manage'),('admin','audit:read');
INSERT INTO role_permissions VALUES('admin','settings:read');
