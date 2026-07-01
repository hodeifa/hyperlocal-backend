-- Migration: 002_create_customers
-- Tabel customers.
-- [REMOVEDv2.3] Kolom refresh_token_hash / refresh_expires_at TIDAK dibuat di
-- sini. Desain final skema ini sudah memakai tabel sesi per-device terpisah
-- (customer_sessions, lihat migrasi 017), jadi kolom tunggal tersebut tidak
-- pernah ada sama sekali pada proyek baru ini (berbeda dari riwayat draf
-- v2.2 yang sempat menaruhnya di sini sebelum direvisi).

CREATE TABLE customers (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_number   VARCHAR(20) NOT NULL UNIQUE,
    name           VARCHAR(100) NOT NULL,
    photo_url      TEXT,
    fcm_token      TEXT,
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    deactivated_at TIMESTAMP,
    created_at     TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP NOT NULL DEFAULT NOW()
);
