-- 000039_debts.up.sql
-- W0 PLAN-BILLING-OPS-V3: hutang-piutang manual (bukan denda/biaya tambahan).
-- Customer dengan hutang tidak dapat dihapus (RESTRICT) agar saldo tidak hilang.

CREATE TABLE customer_debts (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT NOT NULL DEFAULT 'tenant-default',
    customer_id      TEXT NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    subscription_id  TEXT REFERENCES subscriptions(id) ON DELETE SET NULL,
    amount           NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    remaining_amount NUMERIC(15,2) NOT NULL CHECK (remaining_amount >= 0),
    description      TEXT,
    debt_date        DATE NOT NULL,
    due_date         DATE,
    status           VARCHAR(20) NOT NULL DEFAULT 'OPEN',
    paid_at          TIMESTAMPTZ,
    created_by       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_debt_status CHECK (status IN ('OPEN','PARTIAL','PAID','CANCELLED')),
    CONSTRAINT chk_debt_remaining CHECK (remaining_amount <= amount)
);

CREATE INDEX idx_customer_debts_customer ON customer_debts(customer_id, status);
CREATE INDEX idx_customer_debts_due ON customer_debts(due_date);
CREATE TRIGGER trg_customer_debts_upd BEFORE UPDATE ON customer_debts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE debt_payments (
    id                  TEXT PRIMARY KEY,
    tenant_id           TEXT NOT NULL DEFAULT 'tenant-default',
    debt_id             TEXT NOT NULL REFERENCES customer_debts(id) ON DELETE CASCADE,
    amount              NUMERIC(15,2) NOT NULL CHECK (amount > 0),
    payment_method_id   TEXT REFERENCES payment_methods(id) ON DELETE SET NULL,
    cash_transaction_id TEXT REFERENCES cash_transactions(id) ON DELETE SET NULL,
    payment_date        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes               TEXT,
    recorded_by         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_debt_payments_debt ON debt_payments(debt_id, payment_date DESC);

INSERT INTO cash_categories (id, tenant_id, name, type, is_active) VALUES
    ('cc-debt-income', 'tenant-default', 'Pembayaran Hutang', 'INCOME', TRUE)
ON CONFLICT (id) DO NOTHING;

INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.debt_income_category_id', 'cc-debt-income', 'isp_billing',
     'Kategori kas untuk penerimaan pembayaran hutang pelanggan'),
    ('isp.debt_reminder_enabled',   'true',           'isp_billing',
     'Aktifkan pengingat WhatsApp untuk hutang yang mendekati/melewati jatuh tempo'),
    ('isp.debt_reminder_days',      '3,1,0',          'isp_billing',
     'Offset pengingat hutang (hari sebelum jatuh tempo), dipisah koma')
ON CONFLICT (key) DO NOTHING;

INSERT INTO notification_templates (id, template_key, name, content, variables_json) VALUES
    ('nt-debt-reminder', 'DEBT_REMINDER',
     'Pengingat Hutang',
     'Yth {{customer_name}}, terdapat hutang sebesar Rp{{amount}} dengan jatuh tempo {{due_date}}. Mohon segera diselesaikan.',
     '["customer_name","amount","due_date"]'),
    ('nt-debt-paid', 'DEBT_PAID',
     'Hutang Lunas',
     'Terima kasih {{customer_name}}, pembayaran hutang sebesar Rp{{amount}} telah kami terima.',
     '["customer_name","amount"]')
ON CONFLICT (id) DO NOTHING;
