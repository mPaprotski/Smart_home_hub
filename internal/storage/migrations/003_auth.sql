CREATE TABLE bootstrap_state(singleton INTEGER PRIMARY KEY CHECK(singleton=1), initialized_at INTEGER NOT NULL);
CREATE TABLE auth_attempts(account_key TEXT PRIMARY KEY, failures INTEGER NOT NULL CHECK(failures>=0), window_start INTEGER NOT NULL);
CREATE TABLE auth_internal(singleton INTEGER PRIMARY KEY CHECK(singleton=1), limiter_key BLOB NOT NULL CHECK(length(limiter_key)=32));
