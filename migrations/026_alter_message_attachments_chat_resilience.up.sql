-- Migration: 025_alter_message_attachments_chat_resilience
-- [NEWv2.5] Adopsi sequence_diagram_chat.md — Chat Resilience (upload gambar
-- + ClamAV scan + thumbnail async + presigned URL batch refresh).
-- Nomor migrasi digeser dari 017 (draf awal) ke 025 — 017 sudah dipakai
-- customer_sessions (v2.3). Lihat erd.md Section 5 catatan.

-- Buat ENUM attachment_scan_status (konsisten dengan pola verification_status,
-- dispute_status, topup_status — bukan VARCHAR+CHECK)
CREATE TYPE attachment_scan_status AS ENUM (
    'pending',   -- baru diunggah, belum dipindai ClamAV worker
    'clean',     -- lolos pemindaian
    'infected'   -- terdeteksi malware — diblokir dari penayangan
);

-- Tambah kolom v2.5
ALTER TABLE message_attachments
    ADD COLUMN IF NOT EXISTS thumb_url         TEXT,
    ADD COLUMN IF NOT EXISTS thumb_expires_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS scan_status       attachment_scan_status NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS original_filename TEXT,
    ADD COLUMN IF NOT EXISTS etag              VARCHAR(64);

-- Ubah presigned_url_expires_at dari TIMESTAMP ke TIMESTAMPTZ
-- Konversi: asumsikan nilai lama disimpan dalam UTC
ALTER TABLE message_attachments
    ALTER COLUMN presigned_url_expires_at
        TYPE TIMESTAMPTZ USING presigned_url_expires_at AT TIME ZONE 'UTC';

-- Index v2.5: ClamAV worker — hanya baris masih pending (partial index)
CREATE INDEX IF NOT EXISTS idx_message_attachments_scan_status
    ON message_attachments(scan_status)
    WHERE scan_status = 'pending';

-- Index v2.5: Saga compensating transaction — verifikasi etag sebelum DeleteObject
CREATE INDEX IF NOT EXISTS idx_message_attachments_etag
    ON message_attachments(etag)
    WHERE etag IS NOT NULL;
