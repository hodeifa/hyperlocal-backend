-- Migration: 026_create_audit_logs_entity_index
-- [NEWv2.6] Tambah index entity pada audit_logs.
-- Dibutuhkan karena audit_logs sekarang juga diisi untuk setiap wallet.deduct
-- (setiap order COMPLETED) — volume baris naik signifikan, tidak lagi hanya
-- saat anomali top-up yang jarang terjadi (lihat erd.md Section 6).
-- Query "riwayat audit per wallet/entity" tanpa index ini = full-scan tabel
-- yang terus bertumbuh sebesar jumlah order selesai per hari.

CREATE INDEX idx_audit_logs_entity
    ON audit_logs(entity_type, entity_id, created_at DESC);
