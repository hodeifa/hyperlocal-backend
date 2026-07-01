-- Migration: 008_create_orders
-- Tabel orders.
-- [NEWv2.1]  platform_fee, customer_payment_method — sudah inline di sini.
-- [NEWv2.2]  finish_note VARCHAR(150), version (optimistic lock) — sudah inline.
-- Catatan: FK driver_transactions.order_id → orders.id sudah dideklarasikan
-- sebelum tabel ini ada (migrasi 007). PostgreSQL mendukung FK ke-depan via
-- DEFERRABLE atau ALTER TABLE. Karena kita memakai CREATE TABLE dengan FK
-- bersamaan, di sini orders dibuat setelah driver_transactions, sehingga FK
-- seharusnya valid — golang-migrate menjalankan 007 lebih dulu dari 008.
-- Namun FK di 007 ke orders.id baru exist setelah 008 dijalankan.
-- ⚠ Solusi: FK di 007 (driver_transactions.order_id) dibuat dengan DEFERRABLE
-- INITIALLY DEFERRED, atau — lebih sederhana untuk migrasi sequential — FK
-- tersebut ditambahkan via ALTER TABLE setelah orders ada. Untuk kemudahan,
-- FK driver_transactions.order_id ditambahkan di migrasi ini (008) via ALTER.

ALTER TABLE driver_transactions
    ADD CONSTRAINT fk_driver_transactions_order
        FOREIGN KEY (order_id) REFERENCES orders(id);

-- Perhatian: baris ALTER di atas akan GAGAL jika tabel orders belum exist saat
-- 007 dieksekusi. Karena golang-migrate menjalankan 007 → 008 berurutan, dan
-- dalam 007 FK order_id tidak mempunyai REFERENCES orders(id) (hanya UUID biasa),
-- maka FK sebenarnya baru di-enforce di 008. Lihat 007 — kolom order_id di sana
-- dideklarasikan tanpa REFERENCES, lalu di sini kita tambahkan constraint-nya.

CREATE TYPE order_status AS ENUM (
    'SEARCHING',      -- order baru, mencari mitra
    'ACCEPTED',       -- mitra terima, menuju titik jemput
    'IN_PROGRESS',    -- perjalanan berlangsung
    'COMPLETED',      -- selesai, platform_fee sudah dipotong
    'CANCELLED',      -- dibatalkan customer
    'NO_DRIVER_FOUND' -- tidak ada mitra dalam radius setelah timer habis
);

CREATE TYPE customer_payment_method AS ENUM (
    'CASH',
    'BANK_TRANSFER'
);

CREATE TABLE orders (
    id                       UUID          PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id              UUID          NOT NULL REFERENCES customers(id),
    driver_id                UUID          REFERENCES drivers(id),
    status                   order_status  NOT NULL DEFAULT 'SEARCHING',
    pickup_address           TEXT          NOT NULL,
    dropoff_address          TEXT          NOT NULL,
    pickup_point             GEOMETRY(Point, 4326) NOT NULL,
    dropoff_point            GEOMETRY(Point, 4326) NOT NULL,
    polyline                 TEXT,
    distance_meters          INT           NOT NULL,
    fare                     DECIMAL(12,2) NOT NULL,
    platform_fee             DECIMAL(12,2) NOT NULL,
    customer_payment_method  customer_payment_method,
    finish_note              VARCHAR(150),
    rating                   INT           CHECK (rating BETWEEN 1 AND 5),
    rating_comment           TEXT,
    started_at               TIMESTAMP,
    completed_at             TIMESTAMP,
    created_at               TIMESTAMP     NOT NULL DEFAULT NOW(),
    version                  INT           NOT NULL DEFAULT 1
);
