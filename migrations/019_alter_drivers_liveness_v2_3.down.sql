ALTER TABLE drivers
    DROP COLUMN IF EXISTS liveness_retry_count,
    DROP COLUMN IF EXISTS ocr_retry_count,
    DROP COLUMN IF EXISTS liveness_score;
