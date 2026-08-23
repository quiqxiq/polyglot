-- 000015_rebuild_isp_core.down.sql
-- Kembalikan bentuk minimal dari migrasi 000006.

DROP TABLE IF EXISTS invoice_items;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS customer_portal_sessions;
DROP TABLE IF EXISTS customers;

CREATE TABLE customers (
    id         TEXT PRIMARY KEY,
    tenant_id  TEXT,
    name       TEXT,
    email      TEXT,
    phone      TEXT,
    address    TEXT,
    status     TEXT,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
);

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

CREATE TABLE invoices (
    id          TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    amount      NUMERIC NOT NULL,
    status      TEXT NOT NULL DEFAULT 'UNPAID',
    due_date    TIMESTAMPTZ NOT NULL,
    paid_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ
);

CREATE INDEX idx_invoices_customer_id ON invoices (customer_id);
