DROP INDEX IF EXISTS idx_message_attachments_etag;
DROP INDEX IF EXISTS idx_message_attachments_scan_status;

ALTER TABLE message_attachments
    ALTER COLUMN presigned_url_expires_at
        TYPE TIMESTAMP USING presigned_url_expires_at AT TIME ZONE 'UTC';

ALTER TABLE message_attachments
    DROP COLUMN IF EXISTS etag,
    DROP COLUMN IF EXISTS original_filename,
    DROP COLUMN IF EXISTS scan_status,
    DROP COLUMN IF EXISTS thumb_expires_at,
    DROP COLUMN IF EXISTS thumb_url;

DROP TYPE IF EXISTS attachment_scan_status;
