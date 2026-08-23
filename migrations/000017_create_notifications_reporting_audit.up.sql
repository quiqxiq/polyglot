-- 000017_create_notifications_reporting_audit.up.sql
-- Fase 2 (DATABASE-SCHEMA-ISP.md §2.9–2.10): template & log WA,
-- snapshot harian, audit_logs + seed template WA (idempoten).

-- 16. Master Template Pesan WhatsApp
CREATE TABLE notification_templates (
    id             TEXT PRIMARY KEY,
    tenant_id      TEXT NOT NULL DEFAULT 'tenant-default',
    template_key   VARCHAR(50) NOT NULL,
    name           VARCHAR(100) NOT NULL,
    content        TEXT NOT NULL,
    variables_json JSONB NOT NULL DEFAULT '[]',
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_notif_template UNIQUE (tenant_id, template_key)
);

CREATE TRIGGER trg_notif_templates_upd BEFORE UPDATE ON notification_templates
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- 17. Antrean & Log Notifikasi WhatsApp
CREATE TABLE wa_notifications (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL DEFAULT 'tenant-default',
    template_id     TEXT REFERENCES notification_templates(id) ON DELETE SET NULL,
    customer_id     TEXT REFERENCES customers(id) ON DELETE SET NULL,
    invoice_id      TEXT REFERENCES invoices(id) ON DELETE SET NULL,
    recipient_phone VARCHAR(20) NOT NULL,
    message_type    VARCHAR(50) NOT NULL,
    message_content TEXT NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'QUEUED',
    error_message   TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_wa_notif_cust   ON wa_notifications(customer_id, created_at DESC);
CREATE INDEX idx_wa_notif_status ON wa_notifications(status, created_at DESC);

-- 15. Snapshot Rekap Finansial Harian
CREATE TABLE daily_financial_snapshots (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id            TEXT NOT NULL DEFAULT 'tenant-default',
    snapshot_date        DATE NOT NULL,
    invoice_count        INT NOT NULL DEFAULT 0,
    invoice_total        NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    payment_count        INT NOT NULL DEFAULT 0,
    payment_total        NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    outstanding_total    NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    expense_total        NUMERIC(15,2) NOT NULL DEFAULT 0.00,
    active_subscriptions INT NOT NULL DEFAULT 0,
    cash_balance_json    JSONB NOT NULL DEFAULT '{}',
    created_at           TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_daily_snapshot UNIQUE (tenant_id, snapshot_date)
);

CREATE INDEX idx_daily_snapshot_date ON daily_financial_snapshots(snapshot_date DESC);

-- 18. Audit Trail Aktivitas Penting Sistem
CREATE TABLE audit_logs (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id   TEXT NOT NULL DEFAULT 'tenant-default',
    actor_type  VARCHAR(20) NOT NULL DEFAULT 'USER',   -- USER | SYSTEM | PORTAL
    actor_id    TEXT,
    action      VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id   TEXT NOT NULL,
    description TEXT,
    ip_address  VARCHAR(45),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
CREATE INDEX idx_audit_logs_date   ON audit_logs(created_at DESC);

INSERT INTO notification_templates (id, template_key, name, content, variables_json) VALUES
    ('nt-bill-reminder', 'BILL_REMINDER',
     'Pemberitahuan Tagihan Bulanan',
     'Yth {{customer_name}}, tagihan {{period}} sebesar Rp{{total}} telah terbit. Jatuh tempo {{due_date}}. Kode bayar: {{payment_code}}.',
     '["customer_name","period","total","due_date","payment_code"]'),
    ('nt-payment-receipt', 'PAYMENT_RECEIPT',
     'Bukti Pembayaran Lunas',
     'Terima kasih {{customer_name}}, pembayaran Rp{{amount}} untuk tagihan {{period}} telah kami terima pada {{paid_at}}.',
     '["customer_name","amount","period","paid_at"]'),
    ('nt-reg-approved', 'REGISTRATION_APPROVED',
     'Pendaftaran Disetujui',
     'Halo {{full_name}}, pendaftaran paket {{plan_name}} Anda disetujui. Jadwal pemasangan: {{install_schedule}}.',
     '["full_name","plan_name","install_schedule"]'),
    ('nt-install-scheduled', 'INSTALLATION_SCHEDULED',
     'Jadwal Pemasangan',
     'Halo {{full_name}}, teknisi kami akan datang pada {{install_schedule}} ke alamat {{address}}.',
     '["full_name","install_schedule","address"]'),
    ('nt-isolation-notice', 'ISOLATION_NOTICE',
     'Pemberitahuan Isolir',
     'Yth {{customer_name}}, layanan Anda diisolir sementara karena tagihan melewati jatuh tempo. Lunasi untuk reaktivasi otomatis.',
     '["customer_name"]')
ON CONFLICT (id) DO NOTHING;
