-- 000015_rebuild_isp_core.up.sql
-- Fase 2 (DATABASE-SCHEMA-ISP.md §2.3–2.6): rebuild tabel inti ISP —
-- customers (+portal code/GPS/soft delete), customer_portal_sessions,
-- subscriptions (mapping MikroTik, kredensial tersandi), invoices
-- (split nominal + QR/kode bayar), invoice_items.
-- Development stage: DROP + CREATE tanpa migrasi data lama.

DROP TABLE IF EXISTS invoice_items;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS customer_portal_sessions;
DROP TABLE IF EXISTS customers;

-- 4. Master Pelanggan Aktif
CREATE TABLE customers (
    id                 TEXT PRIMARY KEY,
    tenant_id          TEXT NOT NULL DEFAULT 'tenant-default',
    customer_code      VARCHAR(30)  UNIQUE NOT NULL,
    name               VARCHAR(100) NOT NULL,
    phone              VARCHAR(20)  NOT NULL,
    email              VARCHAR(100),
    address            TEXT         NOT NULL,
    latitude           DOUBLE PRECISION,
    longitude          DOUBLE PRECISION,
    portal_access_code VARCHAR(16)  UNIQUE NOT NULL,
    status             VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE',
    notes              TEXT,
    registered_at      DATE         NOT NULL DEFAULT CURRENT_DATE,
    deleted_at         TIMESTAMPTZ,
    created_at         TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_customers_code   ON customers(customer_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_phone  ON customers(phone) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_portal ON customers(portal_access_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_customers_status ON customers(status) WHERE deleted_at IS NULL;
CREATE TRIGGER trg_customers_upd BEFORE UPDATE ON customers FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 5. Sesi Login Portal Mandiri Pelanggan
CREATE TABLE customer_portal_sessions (
    id            TEXT PRIMARY KEY,
    tenant_id     TEXT NOT NULL DEFAULT 'tenant-default',
    customer_id   TEXT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    session_token VARCHAR(128) UNIQUE NOT NULL,
    ip_address    VARCHAR(45),
    user_agent    TEXT,
    expires_at    TIMESTAMPTZ NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_portal_sessions_token ON customer_portal_sessions(session_token);

-- 6. Langganan Pelanggan (Mapping ke Router MikroTik)
CREATE TABLE subscriptions (
    id                   TEXT PRIMARY KEY,
    tenant_id            TEXT NOT NULL DEFAULT 'tenant-default',
    customer_id          TEXT NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    plan_id              TEXT NOT NULL REFERENCES service_plans(id) ON DELETE RESTRICT,
    device_id            UUID REFERENCES devices(id) ON DELETE RESTRICT, -- Router BRAS target
    service_type         VARCHAR(20)  NOT NULL DEFAULT 'PPPOE',

    -- Kredensial & konfigurasi MikroTik. DEVIASI vs dokumen:
    -- remote_password plaintext diganti ciphertext AES-GCM base64 via vault.
    remote_username      VARCHAR(100) NOT NULL,
    remote_password_cipher TEXT     NOT NULL,
    local_address        VARCHAR(45),
    remote_address       VARCHAR(45),
    parent_queue         VARCHAR(50) DEFAULT 'none',
    rate_limit           VARCHAR(100),

    billing_cycle        VARCHAR(20) NOT NULL DEFAULT 'MONTHLY',
    billing_day          INT         NOT NULL DEFAULT 1,
    auto_isolate         BOOLEAN     NOT NULL DEFAULT TRUE,
    isolation_grace_days INT         NOT NULL DEFAULT 3,
    status               VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
    start_date           DATE        NOT NULL DEFAULT CURRENT_DATE,
    end_date             DATE,
    custom_price         NUMERIC(15,2),
    current_period_start TIMESTAMPTZ,
    current_period_end   TIMESTAMPTZ,
    notes                TEXT,
    deleted_at           TIMESTAMPTZ,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_subscriptions_customer ON subscriptions(customer_id);
CREATE INDEX idx_subscriptions_plan     ON subscriptions(plan_id);
CREATE INDEX idx_subscriptions_device   ON subscriptions(device_id);
CREATE INDEX idx_subscriptions_status   ON subscriptions(status) WHERE deleted_at IS NULL;
CREATE TRIGGER trg_subscriptions_upd BEFORE UPDATE ON subscriptions FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 7. Faktur Tagihan Bulanan
CREATE TABLE invoices (
    id                  TEXT PRIMARY KEY,
    tenant_id           TEXT NOT NULL DEFAULT 'tenant-default',
    invoice_number      VARCHAR(50) UNIQUE NOT NULL,
    customer_id         TEXT NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    subscription_id     TEXT REFERENCES subscriptions(id) ON DELETE SET NULL,
    period              VARCHAR(10) NOT NULL,               -- '2026-08'
    subtotal            NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    discount            NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    tax_amount          NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    total               NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    paid_amount         NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    due_date            DATE NOT NULL,
    paid_at             TIMESTAMPTZ,
    status              VARCHAR(20) NOT NULL DEFAULT 'UNPAID',
    qr_payload          VARCHAR(255) UNIQUE NOT NULL,
    manual_payment_code VARCHAR(30)  UNIQUE NOT NULL,
    notes               TEXT,
    cancelled_at        TIMESTAMPTZ,
    cancel_reason       TEXT,
    deleted_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_invoices_customer ON invoices(customer_id);
CREATE INDEX idx_invoices_period   ON invoices(period);
CREATE INDEX idx_invoices_status   ON invoices(status);
CREATE INDEX idx_invoices_due_date ON invoices(due_date);
CREATE TRIGGER trg_invoices_upd BEFORE UPDATE ON invoices FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 8. Rincian Item Baris Tagihan
CREATE TABLE invoice_items (
    id          TEXT PRIMARY KEY,
    invoice_id  TEXT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    description VARCHAR(255) NOT NULL,
    quantity    NUMERIC(12,2) NOT NULL DEFAULT 1.00,
    unit_price  NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    amount      NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    item_type   VARCHAR(30)  NOT NULL DEFAULT 'SUBSCRIPTION_FEE',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_invoice_items_invoice ON invoice_items(invoice_id);
