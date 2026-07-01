-- Migration: 023_alter_outbox_events_for_backoff
-- [NEWv2.4] Tambah next_attempt_at untuk exponential backoff di OutboxWorker.
-- ⚠ PENTING: retry_count dan last_error JANGAN di-ADD COLUMN lagi di sini —
-- keduanya SUDAH ADA sejak migrasi 012 (v2.2). Menambahkan ulang akan error
-- "column already exists". Lihat erd.md Section 5, catatan migration 023.
-- Index idx_outbox_pending diganti dari created_at → next_attempt_at agar
-- worker langsung filter event yang sudah waktunya dicoba (backoff selesai).

ALTER TABLE outbox_events
    ADD COLUMN IF NOT EXISTS next_attempt_at TIMESTAMP NOT NULL DEFAULT NOW();

-- Update existing pending rows: next_attempt_at = created_at (backoff mulai dari awal)
UPDATE outbox_events
SET next_attempt_at = created_at
WHERE processed_at IS NULL
  AND next_attempt_at = NOW(); -- hanya baris yang baru di-set DEFAULT-nya

-- Ganti index lama ke basis next_attempt_at
DROP INDEX IF EXISTS idx_outbox_pending;

CREATE INDEX idx_outbox_pending
    ON outbox_events(next_attempt_at)
    WHERE processed_at IS NULL;
