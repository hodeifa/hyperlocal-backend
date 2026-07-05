-- Migration: 018_create_driver_sessions
-- [NEWv2.3] Tabel driver_sessions — struktur identik customer_sessions.
-- Migrasi data dan DROP kolom lama dari drivers dilakukan hanya jika kolom
-- refresh_token_hash masih ada (sistem yang pakai 004 versi lama).

CREATE TABLE driver_sessions (
    id                  UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id           UUID         NOT NULL REFERENCES drivers(id),
    refresh_token_hash  VARCHAR(64)  NOT NULL UNIQUE,
    device_info         VARCHAR(255),
    created_at          TIMESTAMP    NOT NULL DEFAULT NOW(),
    last_used_at        TIMESTAMP    NOT NULL DEFAULT NOW(),
    expires_at          TIMESTAMP    NOT NULL,
    revoked_at          TIMESTAMP
);

CREATE INDEX idx_driver_sessions_driver_id
    ON driver_sessions(driver_id);

CREATE INDEX idx_driver_sessions_active
    ON driver_sessions(driver_id)
    WHERE revoked_at IS NULL;

CREATE UNIQUE INDEX idx_driver_sessions_refresh_token_hash
    ON driver_sessions(refresh_token_hash);

-- Migrasi data dari kolom lama jika masih ada
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'drivers' AND column_name = 'refresh_token_hash'
    ) THEN
        INSERT INTO driver_sessions (driver_id, refresh_token_hash, expires_at)
        SELECT id, refresh_token_hash, COALESCE(refresh_expires_at, NOW() + INTERVAL '90 days')
        FROM drivers
        WHERE refresh_token_hash IS NOT NULL
        ON CONFLICT DO NOTHING;

        ALTER TABLE drivers DROP COLUMN IF EXISTS refresh_token_hash;
        ALTER TABLE drivers DROP COLUMN IF EXISTS refresh_expires_at;
    END IF;
END;
$$;
