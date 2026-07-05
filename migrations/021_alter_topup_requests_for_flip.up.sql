-- Migration: 021_alter_topup_requests_for_flip
-- [NEWv2.4] Adopsi flip.md — tambah kolom Flip-specific ke driver_topup_requests.
-- Nomor digeser dari 018 (di flip.md) ke 021 karena 018-020 sudah dipakai v2.3.
-- idempotency_key: UUIDv7 dari client (Flutter), mencegah double-tap Create VA.
-- amount_received: nominal bersih dari Flip setelah fee, untuk audit trail.
-- fee: biaya VA yang ditanggung platform (amount - amount_received).

ALTER TABLE driver_topup_requests
    ADD COLUMN IF NOT EXISTS idempotency_key  VARCHAR(64)   UNIQUE,
    ADD COLUMN IF NOT EXISTS amount_received  DECIMAL(12,2),
    ADD COLUMN IF NOT EXISTS fee              DECIMAL(12,2) DEFAULT 0;

-- UPSERT idempotency_key di Saga pattern CREATE VA mengandalkan index ini
CREATE UNIQUE INDEX IF NOT EXISTS idx_driver_topup_requests_idempotency_key
    ON driver_topup_requests(idempotency_key)
    WHERE idempotency_key IS NOT NULL;
