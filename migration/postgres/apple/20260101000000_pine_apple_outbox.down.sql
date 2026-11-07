BEGIN;

DROP INDEX IF EXISTS idx_outbox_unsent;
DROP TABLE IF EXISTS outbox;
DROP INDEX IF EXISTS idx_pine_apple_apple_id;
DROP TABLE IF EXISTS pine_apple;

COMMIT;
