-- Migration: 019_alter_drivers_liveness_v2_3
-- [NEWv2.3] Tambah field liveness dan retry counter ke drivers.
-- roadmap.md Sprint 5: OCR retry maks 3x, liveness retry maks 3x,
-- liveness_score disimpan untuk audit trail verifikasi.

ALTER TABLE drivers
    ADD COLUMN IF NOT EXISTS liveness_score       DECIMAL(5,2),
    ADD COLUMN IF NOT EXISTS ocr_retry_count      INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS liveness_retry_count INT NOT NULL DEFAULT 0;
