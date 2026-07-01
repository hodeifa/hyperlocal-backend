-- ALTER TYPE ADD VALUE tidak bisa di-rollback di PostgreSQL.
-- Restore DEFAULT saja.
ALTER TABLE driver_topup_requests
    ALTER COLUMN status SET DEFAULT 'PENDING';
