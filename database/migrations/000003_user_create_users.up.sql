CREATE TABLE IF NOT EXISTS users (
  id            BIGSERIAL PRIMARY KEY,

  email         TEXT NOT NULL UNIQUE,
  first_name    TEXT NOT NULL,
  last_name     TEXT NOT NULL,
  middle_name   TEXT NULL,

  iin           TEXT NOT NULL UNIQUE,
  phone         TEXT NULL,

  password      TEXT NOT NULL,  -- password hash

  role_id       BIGINT NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,

  session_token TEXT NULL,
  last_seen     TIMESTAMPTZ NULL,

  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- индекс под частые запросы
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_last_seen ON users(last_seen);
