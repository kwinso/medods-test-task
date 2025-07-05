CREATE TABLE
  auths (
    id SERIAL PRIMARY KEY,
    guid VARCHAR(36) NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL UNIQUE,
    ip_address INET NOT NULL,
    user_agent TEXT NOT NULL,
    refreshed_at TIMESTAMP NOT NULL DEFAULT NOW (),
    created_at TIMESTAMP NOT NULL DEFAULT NOW ()
  );