-- Migration: 016_alter_orders_v2_2
-- [NEWv2.2] Tambah finish_note ke tabel orders.
-- IF NOT EXISTS: aman jika 008 sudah menyertakan kolom ini (sistem baru).

ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS finish_note VARCHAR(150);
