-- Migration: 017_create_customer_sessions
-- [NEWv2.3] Tabel customer_sessions — multi-device login.
-- Menggantikan kolom tunggal refresh_token_hash/refresh_expires_at di customers.
-- Migrasi data dari customers dilakukan hanya jika kolom lama masih ada
-- (sistem yang menjalankan 002 versi lama yang memiliki kedua kolom tersebut).
-- Sistem baru: customers.refresh_token_hash tidak pernah ada → langsung buat tabel.

CREATE TABLE customer_sessions (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id         UUID         NOT NULL REFERENCES customers(id),
    refresh_token_hash  VARCHAR(64)  NOT NULL UNIQUE,
    device_info         VARCHAR(255),
    created_at          TIMESTAMP    NOT NULL DEFAULT NOW(),
    last_used_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMP    NOT NULL,
    revoked_at          TIMESTAMP
);

-- Indeks untuk customer_sessions
CREATE INDEX idx_customer_sessions_customer_id
    ON customer_sessions(customer_id);

-- Partial — hanya sesi aktif (belum di-revoke)
CREATE INDEX idx_customer_sessions_active
    ON customer_sessions(customer_id)
    WHERE revoked_at IS NULL;

CREATE UNIQUE INDEX idx_customer_sessions_refresh_token_hash
    ON customer_sessions(refresh_token_hash);

-- Migrasi data dari kolom lama jika masih ada
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'customers' AND column_name = 'refresh_token_hash'
    ) THEN
        INSERT INTO customer_sessions (customer_id, refresh_token_hash, expires_at)
        SELECT id, refresh_token_hash, COALESCE(refresh_expires_at, NOW() + INTERVAL '90 days')
        FROM customers
        WHERE refresh_token_hash IS NOT NULL
        ON CONFLICT DO NOTHING;

        ALTER TABLE customers DROP COLUMN IF EXISTS refresh_token_hash;
        ALTER TABLE customers DROP COLUMN IF EXISTS refresh_expires_at;
    END IF;
END;
$$;
