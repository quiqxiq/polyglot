-- 000019_portal_otp_wa_and_settings.up.sql
-- Fase 4: OTP login portal pelanggan, penghitung percobaan kirim WA,
-- dan konfigurasi portal/payment-gateway via system_settings.

-- OTP login portal (kode sekali pakai, hash sha256).
CREATE TABLE IF NOT EXISTS portal_otps (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL DEFAULT 'tenant-default',
    phone       VARCHAR(20) NOT NULL,
    code_hash   VARCHAR(255) NOT NULL,
    purpose     VARCHAR(30) NOT NULL DEFAULT 'PORTAL_LOGIN',
    attempts    INT NOT NULL DEFAULT 0,
    expires_at  TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_portal_otps_phone ON portal_otps(phone, created_at DESC);

-- Penghitung retry worker pengirim WA.
ALTER TABLE wa_notifications
    ADD COLUMN IF NOT EXISTS attempts INT NOT NULL DEFAULT 0;

INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.portal_session_hours', '12', 'isp_portal',
     'Masa berlaku sesi login portal pelanggan (jam)'),
    ('isp.otp_ttl_minutes',      '5',  'isp_portal',
     'Masa berlaku OTP login portal (menit)'),
    ('isp.otp_max_attempts',     '5',  'isp_portal',
     'Maksimum percobaan salah memasukkan OTP per permintaan'),
    ('isp.wa_send_max_retry',    '3',  'isp_notification',
     'Maksimum percobaan ulang pengiriman notifikasi WhatsApp'),
    ('gw.tripay.enabled',        'false','isp_gateway',
     'Aktifkan integrasi Tripay'),
    ('gw.tripay.endpoint',       'https://tripay.co.id/api', 'isp_gateway',
     'Base URL API Tripay (sandbox: https://tripay.co.id/api-sandbox)'),
    ('gw.tripay.merchant_code',  '',   'isp_gateway',
     'Kode merchant Tripay'),
    ('gw.tripay.api_key',        '',   'isp_gateway',
     'API key Tripay (mode production)'),
    ('gw.tripay.private_key',    '',   'isp_gateway',
     'Private key Tripay untuk signature HMAC'),
    ('gw.tripay.channel',        'QRIS','isp_gateway',
     'Channel pembayaran default (QRIS/BRIVA/BCAVA/dll)'),
    ('gw.tripay.cash_account_id','ca-1001-kas-kantor','isp_gateway',
     'Rekening kas tujuan saat pembayaran gateway diterima'),
    ('gw.tripay.income_category_id','cc-tagihan','isp_gateway',
     'Kategori kas IN untuk pembayaran gateway')
ON CONFLICT (key) DO NOTHING;
