-- Migration: 011_create_indexes
-- Semua indeks untuk tabel 002–010.

-- ============================================================
-- SPASIAL — orders dan customer_addresses
-- ============================================================
CREATE INDEX idx_orders_pickup_point
    ON orders USING GIST(pickup_point);

CREATE INDEX idx_orders_dropoff_point
    ON orders USING GIST(dropoff_point);

CREATE INDEX idx_customer_addresses_point
    ON customer_addresses USING GIST(point);

-- ============================================================
-- ORDERS
-- ============================================================
CREATE INDEX idx_orders_status
    ON orders(status);

CREATE INDEX idx_orders_customer_id
    ON orders(customer_id);

CREATE INDEX idx_orders_driver_id
    ON orders(driver_id);

CREATE INDEX idx_orders_created_at
    ON orders(created_at DESC);

-- [FIXv2.2] Composite — riwayat order customer dengan filter status
CREATE INDEX idx_orders_customer_history
    ON orders(customer_id, status, created_at DESC);

-- [FIXv2.2] Composite partial — riwayat order mitra
CREATE INDEX idx_orders_driver_history
    ON orders(driver_id, status, created_at DESC)
    WHERE driver_id IS NOT NULL;

-- ============================================================
-- DRIVERS
-- ============================================================
CREATE INDEX idx_drivers_status
    ON drivers(status);

-- ============================================================
-- CHAT — messages dan message_attachments
-- ============================================================
CREATE INDEX idx_messages_order_id
    ON messages(order_id);

CREATE INDEX idx_messages_created_at
    ON messages(created_at DESC);

CREATE INDEX idx_message_attachments_message_id
    ON message_attachments(message_id);

-- Index untuk job refresh presigned URL yang mendekati expired
CREATE INDEX idx_message_attachments_expires_at
    ON message_attachments(presigned_url_expires_at);

-- ============================================================
-- WALLET — driver_transactions
-- ============================================================
CREATE INDEX idx_driver_transactions_driver_id
    ON driver_transactions(driver_id);

CREATE INDEX idx_driver_transactions_created_at
    ON driver_transactions(created_at DESC);

-- [FIXv2.1] Trace DEDUCTION → order (partial — hanya baris dengan order)
CREATE INDEX idx_driver_transactions_order_id
    ON driver_transactions(order_id)
    WHERE order_id IS NOT NULL;

-- [FIXv2.1] Trace TOP_UP → topup_request (partial)
CREATE INDEX idx_driver_transactions_topup_request_id
    ON driver_transactions(topup_request_id)
    WHERE topup_request_id IS NOT NULL;

-- [FIXv2.2] Safety net: satu order hanya boleh punya satu DEDUCTION
CREATE UNIQUE INDEX idx_driver_transactions_order_deduction
    ON driver_transactions(order_id)
    WHERE type = 'DEDUCTION' AND order_id IS NOT NULL;

-- ============================================================
-- TOP-UP REQUESTS — driver_topup_requests
-- ============================================================
CREATE INDEX idx_driver_topup_requests_driver_id
    ON driver_topup_requests(driver_id);

-- [FIXv2.1] Lookup webhook callback via external_id
CREATE UNIQUE INDEX idx_driver_topup_requests_external_id
    ON driver_topup_requests(external_id);

CREATE INDEX idx_driver_topup_requests_status
    ON driver_topup_requests(status);

-- Partial — hanya PENDING, untuk job cleanup VA expired
CREATE INDEX idx_driver_topup_requests_expires_at
    ON driver_topup_requests(expires_at)
    WHERE status = 'PENDING';
