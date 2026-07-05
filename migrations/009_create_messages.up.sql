-- Migration: 009_create_messages
-- Tabel messages.
-- [FIXv2.2] Kolom sender_id UUID (tanpa FK) TIDAK pernah dibuat di sini —
-- langsung memakai desain v2.2: dua kolom nullable dengan FK eksplisit
-- (customer_sender_id, driver_sender_id) + CHECK constraint referential integrity.
-- Tidak ada kolom sender_id lama yang perlu di-DROP.

CREATE TYPE sender_type AS ENUM ('customer', 'driver');
CREATE TYPE message_type AS ENUM ('text', 'image');

CREATE TABLE messages (
    id                 UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id           UUID         NOT NULL REFERENCES orders(id),
    sender_type        sender_type  NOT NULL,
    customer_sender_id UUID         REFERENCES customers(id),
    driver_sender_id   UUID         REFERENCES drivers(id),
    content            TEXT,
    message_type       message_type NOT NULL,
    created_at         TIMESTAMP    NOT NULL DEFAULT NOW(),
    -- Tepat satu pengirim per baris — dua FK tidak boleh sekaligus NULL atau sekaligus NOT NULL
    CONSTRAINT chk_messages_sender CHECK (
        (customer_sender_id IS NOT NULL AND driver_sender_id IS NULL)
        OR
        (customer_sender_id IS NULL AND driver_sender_id IS NOT NULL)
    )
);
