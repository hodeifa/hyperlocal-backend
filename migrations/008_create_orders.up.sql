-- Migration: 008_create_orders
CREATE TYPE order_status AS ENUM (
    'SEARCHING', 'ACCEPTED', 'IN_PROGRESS', 'COMPLETED', 'CANCELLED', 'NO_DRIVER_FOUND'
);
CREATE TYPE customer_payment_method AS ENUM ('CASH', 'BANK_TRANSFER');

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

-- FK ditambahkan DI SINI, setelah tabel orders sukses dibuat
ALTER TABLE driver_transactions
ADD CONSTRAINT fk_driver_transactions_order
FOREIGN KEY (order_id) REFERENCES orders(id);