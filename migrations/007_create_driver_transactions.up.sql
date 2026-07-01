-- Migration: 007_create_driver_transactions
-- Tabel driver_transactions.
-- [FIXv2.2] Dipindahkan ke 007 agar FK topup_request_id ke driver_topup_requests (006)
-- bisa dibuat — urutan lama (006 transactions, 010 topup_requests) salah secara logis.
-- [CHANGEDv2.1] Tipe transaksi: EARNING/WITHDRAWAL → TOP_UP/DEDUCTION/WITHDRAWAL.
-- [FIXv2.1]    Tambah topup_request_id (FK nullable) untuk audit trail TOP_UP.

CREATE TYPE transaction_type AS ENUM (
    'TOP_UP',      -- mitra menambah deposit via VA — topup_request_id wajib diisi
    'DEDUCTION',   -- platform potong Rp 1.000 per order COMPLETED — order_id wajib diisi
    'WITHDRAWAL'   -- mitra tarik saldo ke rekening — kedua FK NULL
);

CREATE TABLE driver_transactions (
    id                UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    driver_id         UUID          NOT NULL REFERENCES drivers(id),
    order_id          UUID          REFERENCES orders(id),          -- FK ke orders — diisi jika DEDUCTION
    topup_request_id  UUID          REFERENCES driver_topup_requests(id),  -- diisi jika TOP_UP
    amount            DECIMAL(12,2) NOT NULL CHECK (amount > 0),
    type              transaction_type NOT NULL,
    created_at        TIMESTAMP     NOT NULL DEFAULT NOW()
);
