-- 000016_create_registrations_financial.up.sql
-- Fase 2 (DATABASE-SCHEMA-ISP.md §2.3 & §2.7–2.8): registrations,
-- payments/gateway, dan buku kas. Development stage — tanpa migrasi data lama.

-- 3. Pendaftaran Pelanggan Baru
CREATE TABLE registrations (
    id                     TEXT PRIMARY KEY,
    tenant_id              TEXT NOT NULL DEFAULT 'tenant-default',
    registration_no        VARCHAR(30) UNIQUE NOT NULL,           -- REG-202608-0001
    plan_id                TEXT NOT NULL REFERENCES service_plans(id) ON DELETE RESTRICT,
    full_name              VARCHAR(100) NOT NULL,
    phone                  VARCHAR(20)  NOT NULL,
    email                  VARCHAR(100),
    address                TEXT         NOT NULL,
    latitude               DOUBLE PRECISION,
    longitude              DOUBLE PRECISION,
    notes                  TEXT,
    status                 VARCHAR(20)  NOT NULL DEFAULT 'PENDING',
    reviewed_by            BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at            TIMESTAMPTZ,
    admin_notes            TEXT,
    scheduled_install_date DATE,
    scheduled_install_time TIME,
    assigned_technician_id BIGINT       REFERENCES users(id) ON DELETE SET NULL,
    installed_at           TIMESTAMPTZ,
    technician_notes       TEXT,
    customer_id            TEXT REFERENCES customers(id) ON DELETE SET NULL,
    subscription_id        TEXT REFERENCES subscriptions(id) ON DELETE SET NULL,
    invoice_id             TEXT REFERENCES invoices(id) ON DELETE SET NULL,
    rejected_at            TIMESTAMPTZ,
    rejected_reason        TEXT,
    cancelled_at           TIMESTAMPTZ,
    cancel_reason          TEXT,
    created_at             TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMPTZ  NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_reg_status CHECK (status IN ('PENDING','APPROVED','INSTALLED','ACTIVE','REJECTED','CANCELLED'))
);

CREATE INDEX idx_registrations_status ON registrations(status, created_at DESC);
CREATE INDEX idx_registrations_phone  ON registrations(phone);
CREATE INDEX idx_registrations_tech   ON registrations(assigned_technician_id, status);
CREATE TRIGGER trg_registrations_upd BEFORE UPDATE ON registrations FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 9. Master Metode Pembayaran
CREATE TABLE payment_methods (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL DEFAULT 'tenant-default',
    name        VARCHAR(50) NOT NULL,
    type        VARCHAR(30) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 10. Kwitansi Pembayaran Tagihan
CREATE TABLE payments (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL DEFAULT 'tenant-default',
    payment_no        VARCHAR(50)   UNIQUE NOT NULL,
    invoice_id        TEXT NOT NULL REFERENCES invoices(id) ON DELETE RESTRICT,
    payment_method_id TEXT REFERENCES payment_methods(id) ON DELETE SET NULL,
    amount            NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    payment_date      TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    received_by       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    scan_method       VARCHAR(30)   NOT NULL DEFAULT 'MANUAL',
    reference         VARCHAR(100),
    notes             TEXT,
    created_at        TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_payments_invoice ON payments(invoice_id);
CREATE INDEX idx_payments_date    ON payments(payment_date);

-- 11. Transaksi Payment Gateway Online
CREATE TABLE gateway_transactions (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL DEFAULT 'tenant-default',
    gateway         VARCHAR(30)   NOT NULL,
    external_id     VARCHAR(100)  NOT NULL,
    invoice_id      TEXT REFERENCES invoices(id) ON DELETE CASCADE,
    payment_id      TEXT REFERENCES payments(id) ON DELETE SET NULL,
    amount          NUMERIC(15,2) NOT NULL,
    fee_amount      NUMERIC(15,2) DEFAULT 0.00,
    status          VARCHAR(30)   NOT NULL DEFAULT 'PENDING',
    payment_channel VARCHAR(50),
    payment_url     TEXT,
    qr_string       TEXT,
    raw_callback    JSONB,
    callback_count  INT           NOT NULL DEFAULT 0,
    paid_at         TIMESTAMPTZ,
    expires_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_gateway_ext UNIQUE (gateway, external_id)
);

CREATE INDEX idx_gateway_tx_invoice ON gateway_transactions(invoice_id);
CREATE INDEX idx_gateway_tx_status  ON gateway_transactions(status);
CREATE TRIGGER trg_gateway_tx_upd BEFORE UPDATE ON gateway_transactions FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 12. Rekening Kas & Bank
CREATE TABLE cash_accounts (
    id           TEXT PRIMARY KEY,
    tenant_id    TEXT NOT NULL DEFAULT 'tenant-default',
    account_code VARCHAR(30) UNIQUE NOT NULL,
    name         VARCHAR(100) NOT NULL,
    type         VARCHAR(30) NOT NULL DEFAULT 'CASH',
    is_active    BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- 13. Kategori Pos Arus Kas
CREATE TABLE cash_categories (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL DEFAULT 'tenant-default',
    name        VARCHAR(100) NOT NULL,
    type        VARCHAR(20) NOT NULL,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    CONSTRAINT uq_cash_cat_tenant UNIQUE (tenant_id, name)
);

-- 14. Jurnal Mutasi Arus Kas
CREATE TABLE cash_transactions (
    id             TEXT PRIMARY KEY,
    tenant_id      TEXT NOT NULL DEFAULT 'tenant-default',
    transaction_no VARCHAR(50)   UNIQUE NOT NULL,
    account_id     TEXT NOT NULL REFERENCES cash_accounts(id) ON DELETE RESTRICT,
    category_id    TEXT NOT NULL REFERENCES cash_categories(id) ON DELETE RESTRICT,
    direction      VARCHAR(10)   NOT NULL,
    amount         NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    trx_date       TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    source_type    VARCHAR(30),
    source_id      TEXT,
    description    TEXT NOT NULL,
    recorded_by    BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_cash_direction CHECK (direction IN ('IN', 'OUT'))
);

CREATE INDEX idx_cash_trx_account ON cash_transactions(account_id, trx_date DESC);
CREATE INDEX idx_cash_trx_cat     ON cash_transactions(category_id);
CREATE INDEX idx_cash_trx_date    ON cash_transactions(trx_date DESC);
CREATE INDEX idx_cash_trx_source  ON cash_transactions(source_type, source_id);

INSERT INTO payment_methods (id, name, type) VALUES
    ('pm-cash',    'TUNAI',        'CASH'),
    ('pm-bank',    'TRANSFER_BCA', 'BANK'),
    ('pm-qris',    'QRIS',         'QRIS'),
    ('pm-gateway', 'GATEWAY',      'GATEWAY')
ON CONFLICT (id) DO NOTHING;

INSERT INTO cash_accounts (id, account_code, name, type) VALUES
    ('ca-1001-kas-kantor', '1001-KAS-KANTOR', 'Kas Kasir Utama', 'CASH')
ON CONFLICT (id) DO NOTHING;

INSERT INTO cash_categories (id, name, type) VALUES
    ('cc-tagihan',   'Tagihan Pelanggan', 'INCOME'),
    ('cc-listrik',   'Biaya Listrik',     'EXPENSE'),
    ('cc-bandwidth', 'Bandwidth Uplink',  'EXPENSE'),
    ('cc-gaji',      'Gaji',              'EXPENSE')
ON CONFLICT (id) DO NOTHING;
