-- Migration: 010_create_message_attachments
-- Tabel message_attachments — state awal (sebelum v2.5).
-- Kolom v2.5 (thumb_url, thumb_expires_at, scan_status, original_filename, etag)
-- ditambahkan di migrasi 025.
-- presigned_url_expires_at masih TIMESTAMP di sini — diubah ke TIMESTAMPTZ di 025.

CREATE TABLE message_attachments (
    id                        UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id                UUID         NOT NULL REFERENCES messages(id),
    file_path                 TEXT         NOT NULL,
    presigned_url             TEXT         NOT NULL,
    presigned_url_expires_at  TIMESTAMP    NOT NULL,
    file_size_bytes           INT          NOT NULL,
    mime_type                 VARCHAR(100) NOT NULL,
    created_at                TIMESTAMP    NOT NULL DEFAULT NOW()
);
