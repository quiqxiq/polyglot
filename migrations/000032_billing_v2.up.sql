-- 000032_billing_v2.up.sql
-- Sprint 1 (integrasi-router-billing-psb-global-smartapp.md §3):
-- billing v2 (jenis faktur, prorata, diskon), PSB v2 (survey + instalasi),
-- bukti transfer manual (payment_submissions), audit timeline PSB
-- (registration_events), dan seed system_settings baru.

-- §3.1 service_plans: profil router immutable + mode prorata/pajak + versi.
ALTER TABLE service_plans
    ADD COLUMN IF NOT EXISTS router_profile_name VARCHAR(100),
    ADD COLUMN IF NOT EXISTS prorate_mode VARCHAR(20) NOT NULL DEFAULT 'CALENDAR_DAY',
    ADD COLUMN IF NOT EXISTS tax_mode VARCHAR(20) NOT NULL DEFAULT 'EXCLUSIVE',
    ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

CREATE UNIQUE INDEX IF NOT EXISTS idx_service_plans_profile
    ON service_plans(tenant_id, router_profile_name)
    WHERE router_profile_name IS NOT NULL;

-- §3.2 subscriptions: anchor tagihan per langganan + jejak provisi.
ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS billing_anchor_day INT NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS paid_through DATE,
    ADD COLUMN IF NOT EXISTS next_due_date DATE,
    ADD COLUMN IF NOT EXISTS last_provision_error TEXT,
    ADD COLUMN IF NOT EXISTS provision_attempts INT NOT NULL DEFAULT 0;

-- §3.3 invoices: kolom periode, prorata, diskon, promo, credit note, dan
-- issued_at. Kolom subtotal/discount/tax_amount/total/paid_amount/due_date/
-- paid_at/status/qr_payload/manual_payment_code/cancelled_at/cancel_reason/
-- deleted_at sudah ada sejak 000015 — tidak diduplikasi.
ALTER TABLE invoices
    ADD COLUMN IF NOT EXISTS invoice_type VARCHAR(30) NOT NULL DEFAULT 'RECURRING',
    ADD COLUMN IF NOT EXISTS period_start DATE,
    ADD COLUMN IF NOT EXISTS period_end DATE,
    ADD COLUMN IF NOT EXISTS prorated_days INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS period_days INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS discount_type VARCHAR(20) NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS discount_value NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS promo_code VARCHAR(50),
    ADD COLUMN IF NOT EXISTS credit_note_of TEXT REFERENCES invoices(id),
    ADD COLUMN IF NOT EXISTS issued_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Idempotensi penerbitan tagihan per (subscription_id, period, invoice_type).
ALTER TABLE invoices
    DROP CONSTRAINT IF EXISTS uq_invoices_sub_period_type,
    ADD CONSTRAINT uq_invoices_sub_period_type UNIQUE (subscription_id, period, invoice_type);

-- §3.6 registrations: state machine PSB v2 (superset status lama di 000016,
-- nama constraint tetap chk_reg_status) + kolom survey/instalasi.
ALTER TABLE registrations
    DROP CONSTRAINT IF EXISTS chk_reg_status;

ALTER TABLE registrations
    ADD CONSTRAINT chk_reg_status CHECK (status IN ('PENDING','SURVEY','APPROVED','ON_PROGRESS','INSTALLED','ACTIVE','REJECTED','CANCELLED')),
    ADD COLUMN IF NOT EXISTS nik VARCHAR(20),
    ADD COLUMN IF NOT EXISTS id_card_photo_url TEXT,
    ADD COLUMN IF NOT EXISTS house_photo_url TEXT,
    ADD COLUMN IF NOT EXISTS survey_date DATE,
    ADD COLUMN IF NOT EXISTS survey_time TIME,
    ADD COLUMN IF NOT EXISTS surveyor_id BIGINT REFERENCES users(id),
    ADD COLUMN IF NOT EXISTS survey_status VARCHAR(20),
    ADD COLUMN IF NOT EXISTS survey_notes TEXT,
    ADD COLUMN IF NOT EXISTS estimated_cable_length_meter INT,
    ADD COLUMN IF NOT EXISTS public_token VARCHAR(64) UNIQUE,
    ADD COLUMN IF NOT EXISTS installation_started_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS installation_photo_url TEXT,
    ADD COLUMN IF NOT EXISTS ont_mac_address VARCHAR(30),
    ADD COLUMN IF NOT EXISTS ont_serial_number VARCHAR(50),
    ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    ADD COLUMN IF NOT EXISTS sales_agent_id BIGINT REFERENCES users(id);

-- §3.4 payment_submissions: bukti transfer manual menunggu approval finance.
CREATE TABLE payment_submissions (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL DEFAULT 'tenant-default',
    invoice_id        TEXT NOT NULL REFERENCES invoices(id) ON DELETE CASCADE,
    amount            NUMERIC(15,2) NOT NULL,
    bank_account_id   TEXT REFERENCES cash_accounts(id),
    sender_bank       VARCHAR(50),
    sender_name       VARCHAR(100),
    proof_image_url   TEXT NOT NULL,
    notes             TEXT,
    status            VARCHAR(20) NOT NULL DEFAULT 'PENDING', -- PENDING, APPROVED, REJECTED
    reviewed_by       BIGINT REFERENCES users(id),
    reviewed_at       TIMESTAMPTZ,
    rejection_reason  TEXT,
    payment_id        TEXT REFERENCES payments(id),
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_pay_sub_status ON payment_submissions(status);
CREATE INDEX idx_pay_sub_inv ON payment_submissions(invoice_id);

-- §3.7 registration_events: audit & timeline perjalanan PSB.
CREATE TABLE registration_events (
    id              TEXT PRIMARY KEY,
    registration_id TEXT NOT NULL REFERENCES registrations(id) ON DELETE CASCADE,
    event_type      VARCHAR(50) NOT NULL,
    actor_id        BIGINT REFERENCES users(id),
    title           VARCHAR(100) NOT NULL,
    description     TEXT,
    metadata        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_reg_events_reg ON registration_events(registration_id, created_at ASC);

-- §3.8 seed konfigurasi billing v2 (kategori konsisten dengan 000018:
-- isp_billing; gateway memakai kategori isp_gateway seperti 000028).
-- Anchor siklus tetap memakai key eksisting isp.billing_cycle_start_day.
INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.first_invoice_due_days', '7',        'isp_billing',
     'Jatuh tempo faktur pasang baru dalam jumlah hari kalender'),
    ('isp.bill_isolated',          'true',     'isp_billing',
     'Terbitkan tagihan bagi pelanggan berstatus ISOLATED'),
    ('isp.bill_suspended',         'false',    'isp_billing',
     'Terbitkan tagihan bagi pelanggan berstatus SUSPENDED'),
    ('isp.gateway_fee_bearer',     'MERCHANT', 'isp_gateway',
     'Pihak penanggung biaya gateway (MERCHANT, CUSTOMER, SPLIT)'),
    ('isp.gateway_markup_amount',  '0',        'isp_gateway',
     'Nominal markup komisi admin mitra')
ON CONFLICT (key) DO NOTHING;
