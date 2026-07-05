-- migrations/024_alter_topup_status_enum.down.sql
-- ALTER TYPE ... ADD VALUE tidak bisa di-rollback di PostgreSQL.
SELECT 1;