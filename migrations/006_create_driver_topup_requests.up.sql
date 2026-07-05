-- Migration: 006_create_driver_topup_requests
-- Tabel driver_topup_requests — [NEWv2.1].
-- Ini adalah state v2.1: kolom idempotency_key, amount_received, fee BELUM ada
-- (ditambahkan via ALTER di migrasi 021 dari adopsi flip.md v2.4).
-- ENUM topup_status versi awal (4 nilai): PENDING, PAID, EXPIRED, CANCELLED.
-- Nilai INIT, FAILED, UNDERPAID, OVERPAID ditambahkan via ALTER TYPE di migrasi 024.
-- DEFAULT status = 'PENDING' diubah ke 'INIT' di migrasi 024 setelah INIT tersedia.

CREATE TYPE topup_status AS ENUM (
    'PENDING',    -- VA dibuat, menunggu pembayaran
    'PAID',       -- pembayaran terkonfirmasi dari webhook
    'EXPIRED',    -- melewati expires_at tanpa pembayaran
    'CANCELLED'   -- dibatalkan manual oleh mitra
);

CREATE TABLE driver_topup_requests (
    id                     UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id              UUID          NOT NULL REFERENCES drivers(id),
    external_id            VARCHAR(100)  NOT NULL UNIQUE,
    amount                 DECIMAL(12,2) NOT NULL CHECK (amount > 0),
    virtual_account_number VARCHAR(50)   NOT NULL,
    bank_name              VARCHAR(100)  NOT NULL,
    status                 topup_status  NOT NULL DEFAULT 'PENDING',
    expires_at             TIMESTAMP     NOT NULL,
    paid_at                TIMESTAMP,
    created_at             TIMESTAMP     NOT NULL DEFAULT NOW()
);
