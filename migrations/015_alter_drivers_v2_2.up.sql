-- Migration: 015_alter_drivers_v2_2
-- [NEWv2.2] Kolom-kolom v2.2 di drivers.
-- Karena migrasi 004 sudah dibuat dalam state "bersih" (verification_status,
-- consent_json, dll. sudah inline di CREATE TABLE), semua ADD COLUMN di sini
-- menggunakan IF NOT EXISTS agar aman di sistem baru maupun sistem yang masih
-- menjalankan 004 versi lama (dengan is_verified BOOLEAN).

ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS verification_status          verification_status NOT NULL DEFAULT 'PENDING',
    ADD COLUMN IF NOT EXISTS verification_rejection_reason TEXT,
    ADD COLUMN IF NOT EXISTS consent_json                 JSONB,
    ADD COLUMN IF NOT EXISTS consent_at                   TIMESTAMP,
    ADD COLUMN IF NOT EXISTS auto_decline_count           INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS total_completed_orders       INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS bank_account_encrypted       BYTEA,
    ADD COLUMN IF NOT EXISTS bank_account_last4           CHAR(4),
    ADD COLUMN IF NOT EXISTS is_active                    BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS deactivated_at               TIMESTAMP;

-- Migrasi data: jika kolom is_verified lama masih ada (sistem yang pakai 004 lama),
-- set verification_status = 'APPROVED' untuk driver yang sudah terverifikasi.
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'drivers' AND column_name = 'is_verified'
    ) THEN
        UPDATE drivers SET verification_status = 'APPROVED' WHERE is_verified = TRUE;
        ALTER TABLE drivers DROP COLUMN is_verified;
    END IF;
END;
$$;
