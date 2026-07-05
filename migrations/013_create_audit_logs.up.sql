-- Migration: 013_create_audit_logs
-- Audit trail untuk semua operasi finansial dan sensitif — [FIXv2.2].
-- Index idx_audit_logs_entity ditambahkan di migrasi 026 (v2.6) — saat tabel
-- ini mulai diisi volume tinggi (setiap wallet.deduct per order COMPLETED),
-- bukan hanya anomali top-up yang jarang.

CREATE TABLE audit_logs (
    id          UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_type  VARCHAR(20)  NOT NULL,
    actor_id    UUID,
    action      VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id   UUID,
    metadata    JSONB,
    ip_address  INET,
    created_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);
