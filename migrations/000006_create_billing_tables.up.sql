-- 000006_create_billing_tables.up.sql
-- Tabel billing yang sebelumnya hanya dibuat AutoMigrate dari domain structs.
-- Skema mengikuti definisi model kanonik yang punya gorm tags
-- (subscription.Subscription, billing.Invoice, billing.Plan) serta
-- customer.Customer (tanpa tags) — supaya AutoMigrate dev dan migrasi prod
-- konvergen ke bentuk yang sama.

CREATE TABLE IF NOT EXISTS customers (
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

CREATE TABLE IF NOT EXISTS subscriptions (
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

CREATE INDEX IF NOT EXISTS idx_subscriptions_customer_id ON subscriptions (customer_id);

CREATE TABLE IF NOT EXISTS invoices (
    id          TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    amount      NUMERIC NOT NULL,
    status      TEXT NOT NULL DEFAULT 'UNPAID',
    due_date    TIMESTAMPTZ NOT NULL,
    paid_at     TIMESTAMPTZ,
    created_at  TIMESTAMPTZ,
    updated_at  TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_invoices_customer_id ON invoices (customer_id);

CREATE TABLE IF NOT EXISTS plans (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    speed_mbps  BIGINT NOT NULL,
    price       NUMERIC NOT NULL,
    description TEXT
);
