-- Migration: 005_create_driver_wallets
-- Tabel driver_wallets.
-- [CHANGEDv2.1] Wallet adalah deposit/escrow, bukan earnings wallet.
-- [FIXv2.2]    Batas minimum saldo negatif: Rp -100.000 (100× platform_fee).
-- Trigger set_updated_at() didefinisikan di migrasi 003 dan dipakai di sini
-- agar setiap UPDATE balance otomatis memperbarui updated_at (DEFAULT NOW()
-- hanya berlaku saat INSERT, tidak saat UPDATE — lihat erd.md Section 6).

CREATE TABLE driver_wallets (
    id         UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id  UUID         NOT NULL UNIQUE REFERENCES drivers(id),
    balance    DECIMAL(12,2) NOT NULL DEFAULT 0 CHECK (balance >= -100000),
    updated_at TIMESTAMP    NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trg_driver_wallets_updated_at
BEFORE UPDATE ON driver_wallets
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
