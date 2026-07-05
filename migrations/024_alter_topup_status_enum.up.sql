-- Migration: 024_alter_topup_status_enum
-- [NEWv2.4] Tambah 4 nilai baru ke topup_status ENUM dan ubah DEFAULT kolom.
-- ALTER TYPE ADD VALUE tidak bisa di-rollback di PostgreSQL — tidak ada
-- "REMOVE VALUE". Down migration hanya placeholder.
-- IF NOT EXISTS: aman jika nilai sudah ada (misalnya 006 dibuat di state bersih).

ALTER TYPE topup_status ADD VALUE IF NOT EXISTS 'INIT';
ALTER TYPE topup_status ADD VALUE IF NOT EXISTS 'FAILED';
ALTER TYPE topup_status ADD VALUE IF NOT EXISTS 'UNDERPAID';
ALTER TYPE topup_status ADD VALUE IF NOT EXISTS 'OVERPAID';