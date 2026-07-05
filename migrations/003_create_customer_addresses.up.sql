-- Migration: 003_create_customer_addresses
-- Tabel customer_addresses + fungsi trigger set_updated_at(), dipakai bersama
-- oleh tabel ini, drivers (004), dan driver_wallets (005) -- lihat erd.md
-- Section 6 "Mengapa driver_wallets.updated_at butuh trigger".
-- Didefinisikan di sini (bukan di 005) karena ini migrasi pertama yang
-- membutuhkannya secara berurutan.

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TABLE customer_addresses (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id UUID NOT NULL REFERENCES customers(id),
    label       VARCHAR(50) NOT NULL,
    address     TEXT NOT NULL,
    point       GEOMETRY(Point, 4326) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_customer_addresses_updated_at
BEFORE UPDATE ON customer_addresses
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
