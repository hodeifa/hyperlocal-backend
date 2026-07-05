-- migrations/025_alter_topup_requests_default_init.down.sql
ALTER TABLE driver_topup_requests
    ALTER COLUMN status SET DEFAULT 'PENDING';