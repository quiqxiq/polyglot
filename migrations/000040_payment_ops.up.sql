-- 000040_payment_ops.up.sql
-- W0 PLAN-BILLING-OPS-V3: master metode pembayaran manual, settlement
-- payment gateway (pencatatan fee & pencairan), dan batch mass payment.
-- CATATAN KEPUTUSAN: tidak ada denda keterlambatan dan tidak ada biaya
-- tambahan/extra fee pada tagihan.

ALTER TABLE payment_methods
    ADD COLUMN IF NOT EXISTS bank_name       VARCHAR(80),
    ADD COLUMN IF NOT EXISTS account_number  VARCHAR(50),
    ADD COLUMN IF NOT EXISTS account_holder  VARCHAR(100),
    ADD COLUMN IF NOT EXISTS qris_image_url  TEXT,
    ADD COLUMN IF NOT EXISTS instructions    TEXT,
    ADD COLUMN IF NOT EXISTS sort_order      INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

CREATE TABLE gateway_settlements (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT NOT NULL DEFAULT 'tenant-default',
    gateway          VARCHAR(30) NOT NULL,
    settlement_ref   VARCHAR(100) NOT NULL,
    period_from      DATE,
    period_to        DATE,
    gross_amount     NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    fee_amount       NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    net_amount       NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    bank_account_id  TEXT REFERENCES cash_accounts(id) ON DELETE SET NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    notes            TEXT,
    created_by       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_gateway_settlement_status CHECK (status IN ('PENDING','MATCHED','DISCREPANCY')),
    CONSTRAINT uq_gateway_settlement UNIQUE (tenant_id, gateway, settlement_ref)
);
CREATE INDEX idx_gateway_settlements_period ON gateway_settlements(gateway, period_from, period_to);
CREATE TRIGGER trg_gateway_settlements_upd BEFORE UPDATE ON gateway_settlements
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE bulk_payment_batches (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL DEFAULT 'tenant-default',
    created_by        BIGINT REFERENCES users(id) ON DELETE SET NULL,
    payment_method_id TEXT REFERENCES payment_methods(id) ON DELETE SET NULL,
    invoice_count     INT NOT NULL DEFAULT 0,
    status            VARCHAR(20) NOT NULL DEFAULT 'QUEUED',
    job_id            TEXT,
    notes             TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_bulk_payment_status CHECK (status IN ('QUEUED','RUNNING','DONE','PARTIAL','FAILED'))
);

CREATE TABLE bulk_payment_items (
    id           TEXT PRIMARY KEY,
    batch_id     TEXT NOT NULL REFERENCES bulk_payment_batches(id) ON DELETE CASCADE,
    invoice_id   TEXT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount       NUMERIC(15,2) NOT NULL,
    status       VARCHAR(20) NOT NULL DEFAULT 'QUEUED',
    error        TEXT,
    payment_id   TEXT REFERENCES payments(id) ON DELETE SET NULL,
    processed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_bulk_payment_item_status CHECK (status IN ('QUEUED','PROCESSING','SUCCEEDED','FAILED'))
);
CREATE INDEX idx_bulk_payment_items_batch ON bulk_payment_items(batch_id, status);
CREATE INDEX idx_bulk_payment_items_invoice ON bulk_payment_items(invoice_id);

INSERT INTO cash_categories (id, tenant_id, name, type, is_active) VALUES
    ('cc-gateway-fee', 'tenant-default', 'Biaya Payment Gateway', 'EXPENSE', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.gateway_fee_expense_category_id', 'cc-gateway-fee', 'isp_gateway',
     'Kategori kas (EXPENSE) untuk biaya payment gateway saat transaksi settled'),
    ('isp.mass_payment_enabled', 'true', 'isp_billing',
     'Aktifkan endpoint mass payment (job queue) untuk pembayaran banyak tagihan')
ON CONFLICT (key) DO NOTHING;

INSERT INTO notification_templates (id, template_key, name, content, variables_json) VALUES
    ('nt-payment-rejected', 'PAYMENT_REJECTED',
     'Bukti Pembayaran Ditolak',
     'Halo {{customer_name}}, bukti pembayaran untuk tagihan {{invoice_number}} ditolak. Alasan: {{reason}}. Silakan unggah bukti yang benar.',
     '["customer_name","invoice_number","reason"]')
ON CONFLICT (id) DO NOTHING;
