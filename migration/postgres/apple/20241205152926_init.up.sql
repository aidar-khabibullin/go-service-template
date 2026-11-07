BEGIN;

CREATE TABLE IF NOT EXISTS apple(
    id     UUID PRIMARY KEY,
    name   TEXT,
    status TEXT,

    created_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at  TIMESTAMPTZ,
    deleted_at  TIMESTAMPTZ
);

CREATE INDEX idx_apple_name ON apple (name);

COMMIT;
