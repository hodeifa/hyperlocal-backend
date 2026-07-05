-- Migration: 020_alter_disputes_photo_v2_3
-- [FIXv2.3] Ganti photo_url (URL flat, kedaluarsa setelah 24 jam) dengan
-- pola path+presigned yang sama seperti message_attachments.
-- Lihat erd.md Section 6 "Mengapa disputes.photo_url diganti".
-- ⚠ Konversi photo_url → photo_path tidak presisi (photo_url lama adalah
-- presigned URL, bukan path MinIO mentah). Untuk dispute lama dalam masa
-- retensi 90 hari, jalankan job sekali-jalan terpisah sebelum DROP COLUMN.

ALTER TABLE disputes
    ADD COLUMN IF NOT EXISTS photo_path                    TEXT,
    ADD COLUMN IF NOT EXISTS photo_presigned_url           TEXT,
    ADD COLUMN IF NOT EXISTS photo_presigned_url_expires_at TIMESTAMP;

-- Best-effort migration: catat photo_url lama ke photo_path sebagai referensi.
-- Nilai ini BUKAN path MinIO valid — hanya pelestarian data untuk reconciliation.
UPDATE disputes
SET photo_path = photo_url
WHERE photo_url IS NOT NULL
  AND photo_path IS NULL;

-- DROP kolom lama setelah data di-copy
ALTER TABLE disputes DROP COLUMN IF EXISTS photo_url;

-- Index untuk job refresh presigned URL foto bukti dispute
CREATE INDEX IF NOT EXISTS idx_disputes_photo_expires_at
    ON disputes(photo_presigned_url_expires_at)
    WHERE photo_path IS NOT NULL;
