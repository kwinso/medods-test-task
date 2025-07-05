CREATE TABLE
  IF NOT EXISTS auths (
    id SERIAL PRIMARY KEY,
    guid UUID NOT NULL,
    refresh_token varchar(255) NOT NULL,
    ip_address inet NOT NULL,
    user_agent text NOT NULL,
    refreshed_at TIMESTAMP NOT NULL DEFAULT NOW (),
    created_at TIMESTAMP NOT NULL DEFAULT NOW ()
  );