-- Separate append-only migration: never change an already applied checksum.
CREATE INDEX habit_retention_time ON habit_observations(occurred_at,id);
CREATE INDEX automation_audit_time ON automation_audit(occurred_at,id);
