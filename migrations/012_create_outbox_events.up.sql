-- Migration: 012_create_outbox_events
-- Transactional Outbox Pattern — [FIXv2.2].
-- Kolom next_attempt_at BELUM ada di sini (ditambahkan di migrasi 023 v2.4).
-- retry_count dan last_error SUDAH ada sejak migrasi ini — JANGAN di-ADD ulang
-- di migrasi 023 (akan error "column already exists").
-- Index awal: berbasis created_at. Diubah ke next_attempt_at di 023.

CREATE TABLE outbox_events (
    id             UUID      PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type TEXT      NOT NULL,
    aggregate_id   UUID      NOT NULL,
    event_type     TEXT      NOT NULL,
    payload        JSONB     NOT NULL,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    processed_at   TIMESTAMP,
    retry_count    INT       NOT NULL DEFAULT 0,
    last_error     TEXT
);

-- Index awal — diubah ke next_attempt_at oleh migrasi 023
CREATE INDEX idx_outbox_pending
    ON outbox_events(created_at)
    WHERE processed_at IS NULL;
