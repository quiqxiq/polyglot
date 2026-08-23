-- 000014_rebuild_service_plans.down.sql
-- Kembalikan bentuk minimal tabel plans dari migrasi 000006.

DROP TABLE IF EXISTS service_plans;

CREATE TABLE plans (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    speed_mbps  BIGINT NOT NULL,
    price       NUMERIC NOT NULL,
    description TEXT
);
