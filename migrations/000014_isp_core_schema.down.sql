-- 000014_isp_core_schema.down.sql

DROP TABLE IF EXISTS registrations CASCADE;
DROP TABLE IF EXISTS secrets CASCADE;
DROP TABLE IF EXISTS subscriptions CASCADE;
DROP TABLE IF EXISTS service_plans CASCADE;

DELETE FROM system_settings WHERE category = 'isolir';

-- Kembalikan bentuk skeleton 000006.
CREATE TABLE subscriptions (
    id          TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    plan_id     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'ACTIVE',
    start_date  TIMESTAMPTZ NOT NULL,
    end_date    TIMESTAMPTZ NOT NULL,
    price       NUMERIC NOT NULL,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ
);
CREATE INDEX idx_subscriptions_customer_id ON subscriptions (customer_id);

CREATE TABLE plans (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    speed_mbps  BIGINT NOT NULL,
    price       NUMERIC NOT NULL,
    description TEXT
);

ALTER TABLE customers
    DROP COLUMN IF EXISTS customer_code,
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS notes,
    DROP COLUMN IF EXISTS deleted_at;
