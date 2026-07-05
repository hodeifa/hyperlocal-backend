DROP INDEX IF EXISTS idx_outbox_pending;
-- Restore index lama berbasis created_at
CREATE INDEX idx_outbox_pending
    ON outbox_events(created_at)
    WHERE processed_at IS NULL;
ALTER TABLE outbox_events DROP COLUMN IF EXISTS next_attempt_at;
