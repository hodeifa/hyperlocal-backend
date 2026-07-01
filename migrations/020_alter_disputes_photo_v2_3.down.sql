DROP INDEX IF EXISTS idx_disputes_photo_expires_at;
ALTER TABLE disputes
    ADD COLUMN IF NOT EXISTS photo_url TEXT;
-- Tidak bisa restore data photo_url lama yang sudah di-DROP.
ALTER TABLE disputes
    DROP COLUMN IF EXISTS photo_presigned_url_expires_at,
    DROP COLUMN IF EXISTS photo_presigned_url,
    DROP COLUMN IF EXISTS photo_path;
