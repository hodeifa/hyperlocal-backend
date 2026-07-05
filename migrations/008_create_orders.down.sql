ALTER TABLE driver_transactions DROP CONSTRAINT IF EXISTS fk_driver_transactions_order;
DROP TABLE IF EXISTS orders;
DROP TYPE IF EXISTS customer_payment_method;
DROP TYPE IF EXISTS order_status;
