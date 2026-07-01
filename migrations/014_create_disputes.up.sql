-- Migration: 014_create_disputes
-- Tabel disputes — [NEWv2.2].
-- State awal v2.2: menyimpan photo_url (URL flat).
-- Migrasi 020 (v2.3) mengganti photo_url dengan photo_path + presigned_url pattern.
-- reporter_type menggunakan ENUM terpisah (nilai identik dengan sender_type tapi
-- konsep berbeda — reporter vs pengirim pesan).

CREATE TYPE dispute_reporter_type AS ENUM ('customer', 'driver');

CREATE TYPE dispute_status AS ENUM (
    'OPEN',      -- baru dibuat, belum ditangani admin
    'IN_REVIEW', -- sedang ditangani admin
    'RESOLVED',  -- selesai, masalah terselesaikan
    'CLOSED'     -- ditutup tanpa penyelesaian
);

CREATE TABLE disputes (
    id            UUID                  PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id      UUID                  NOT NULL REFERENCES orders(id),
    reporter_type dispute_reporter_type NOT NULL,
    reporter_id   UUID                  NOT NULL,
    text          VARCHAR(200)          NOT NULL,
    photo_url     TEXT,
    status        dispute_status        NOT NULL DEFAULT 'OPEN',
    created_at    TIMESTAMP             NOT NULL DEFAULT NOW()
);
