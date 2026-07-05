-- migrations/025_alter_topup_requests_default_init.up.sql
-- Setelah COMMIT migration 024, nilai 'INIT' sudah visible dan aman dipakai.
ALTER TABLE driver_topup_requests
    ALTER COLUMN status SET DEFAULT 'INIT';