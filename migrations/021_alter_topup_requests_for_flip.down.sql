DROP INDEX IF EXISTS idx_driver_topup_requests_idempotency_key;
ALTER TABLE driver_topup_requests
    DROP COLUMN IF EXISTS fee,
    DROP COLUMN IF EXISTS amount_received,
    DROP COLUMN IF EXISTS idempotency_key;
