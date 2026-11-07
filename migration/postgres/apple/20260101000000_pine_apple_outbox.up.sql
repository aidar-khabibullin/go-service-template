BEGIN;

CREATE TABLE IF NOT EXISTS pine_apple(
    id       UUID PRIMARY KEY,
    apple_id UUID NOT NULL REFERENCES apple (id),
    name     TEXT NOT NULL,
    status   TEXT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_pine_apple_apple_id ON pine_apple (apple_id);

-- Transactional outbox: events are written together with business data
-- and delivered to Kafka by a background worker.
CREATE TABLE IF NOT EXISTS outbox(
    id         UUID PRIMARY KEY,
    key        TEXT  NOT NULL,
    payload    BYTEA NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    sent_at    TIMESTAMPTZ
);

CREATE INDEX idx_outbox_unsent ON outbox (created_at) WHERE sent_at IS NULL;

COMMIT;
