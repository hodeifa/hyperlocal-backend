-- Migration: 004_create_drivers
-- Tabel drivers.
-- Catatan versi:
--   [NEWv2.1]  Tambah bank_name, bank_account_encrypted (FIXv2.2 ganti plaintext),
--               bank_account_last4, bank_account_name, total_completed_orders.
--   [NEWv2.2]  Ganti is_verified BOOLEAN → verification_status ENUM (sudah inline
--               di CREATE TABLE ini). Tambah consent_json, consent_at,
--               auto_decline_count, verification_rejection_reason, is_active,
--               deactivated_at.
--   [REMOVEDv2.3] refresh_token_hash / refresh_expires_at TIDAK dibuat di sini —
--               digantikan tabel driver_sessions (lihat migrasi 018).
--   Liveness fields (liveness_score, ocr_retry_count, liveness_retry_count)
--               TIDAK ada di sini — ditambahkan via ALTER di migrasi 019.

CREATE TYPE vehicle_type AS ENUM ('BICYCLE', 'MOTORCYCLE', 'ON_FOOT');
CREATE TYPE driver_status AS ENUM ('ONLINE', 'BUSY', 'OFFLINE');
CREATE TYPE verification_status AS ENUM ('PENDING', 'APPROVED', 'REJECTED', 'ESCALATED');

CREATE TABLE drivers (
    id                           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number                 VARCHAR(20) NOT NULL UNIQUE,
    name                         VARCHAR(100) NOT NULL,
    photo_url                    TEXT,
    vehicle_type                 vehicle_type NOT NULL,
    ktp_photo_url                TEXT,
    vehicle_photo_url            TEXT,
    verification_status          verification_status NOT NULL DEFAULT 'PENDING',
    verification_rejection_reason TEXT,
    consent_json                 JSONB,
    consent_at                   TIMESTAMP,
    auto_decline_count           INT         NOT NULL DEFAULT 0,
    status                       driver_status NOT NULL DEFAULT 'OFFLINE',
    rating                       DECIMAL(3,2),
    total_completed_orders       INT         NOT NULL DEFAULT 0,
    fcm_token                    TEXT,
    bank_name                    VARCHAR(100),
    bank_account_encrypted       BYTEA,
    bank_account_last4           CHAR(4),
    bank_account_name            VARCHAR(100),
    is_active                    BOOLEAN     NOT NULL DEFAULT TRUE,
    deactivated_at               TIMESTAMP,
    created_at                   TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at                   TIMESTAMP   NOT NULL DEFAULT NOW()
);

-- Trigger set_updated_at() sudah didefinisikan di migrasi 003.
CREATE TRIGGER trg_drivers_updated_at
BEFORE UPDATE ON drivers
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
