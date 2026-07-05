-- Migration: 001_create_extensions
-- Sprint 2 - Hari 1 (roadmap.md). Aktifkan ekstensi PostgreSQL yang dibutuhkan skema ini.
--   pgcrypto -> gen_random_uuid() (semua PK UUID) + pgp_sym_encrypt/decrypt
--              (drivers.bank_account_encrypted, lihat erd.md Section 6)
--   postgis  -> tipe GEOMETRY(Point, 4326) (orders.pickup_point/dropoff_point,
--              customer_addresses.point)

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS postgis;
