-- Migration: 022_create_processed_flip_webhooks
-- [NEWv2.4] Anti-replay lapis kedua untuk webhook Flip.
-- Flip retry webhook yang belum dapat HTTP 200 tiap ~1 menit selama 24 jam.
-- INSERT ... ON CONFLICT (flip_event_id) DO NOTHING di awal handler memastikan
-- satu flip_event_id hanya pernah diproses sekali, terlepas dari guard status
-- yang ada di driver_topup_requests.
-- bill_id mengacu ke driver_topup_requests.external_id (tidak FK formal karena
-- bisa saja external_id belum ada jika Flip kirim event sebelum PENDING tercatat).

CREATE TABLE processed_flip_webhooks (
    flip_event_id VARCHAR(100) PRIMARY KEY,
    bill_id       VARCHAR(100) NOT NULL,
    status        VARCHAR(20)  NOT NULL,
    processed_at  TIMESTAMP    NOT NULL DEFAULT NOW()
);

-- Lookup by bill_id (untuk debug: lihat semua event Flip untuk satu topup)
CREATE INDEX idx_processed_flip_webhooks_bill_id
    ON processed_flip_webhooks(bill_id);
