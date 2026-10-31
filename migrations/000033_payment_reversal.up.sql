-- 000033_payment_reversal.up.sql
-- Fase 3 (integrasi-router-billing-psb-global-smartapp.md §4 Fase 3):
-- reversal pembayaran (void + jurnal koreksi kas) dan konfigurasi default
-- approval bukti transfer manual + estimasi biaya gateway.

-- Kolom void pada payments: baris pembayaran tetap ada untuk audit trail,
-- koreksi dicatat sebagai transaksi kas keluar (cc-refund).
ALTER TABLE payments
    ADD COLUMN IF NOT EXISTS voided_at   TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS void_reason TEXT,
    ADD COLUMN IF NOT EXISTS voided_by   BIGINT REFERENCES users(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_payments_voided
    ON payments(voided_at) WHERE voided_at IS NOT NULL;

-- Default approval bukti transfer manual (dipakai PaymentSubmissionUseCase).
INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.manual_payment_cash_account_id',   'ca-1001-kas-kantor', 'isp_billing',
     'Rekening kas tujuan saat menyetujui bukti transfer manual'),
    ('isp.manual_payment_income_category_id', 'cc-tagihan',        'isp_billing',
     'Kategori kas IN saat menyetujui bukti transfer manual'),
    ('isp.manual_payment_method_id',          'pm-bank',           'isp_billing',
     'Metode pembayaran saat menyetujui bukti transfer manual'),
    ('isp.gateway_fee_percent',               '0',                 'isp_gateway',
     'Estimasi biaya gateway yang dibebankan ke pelanggan (persen dari tagihan)')
ON CONFLICT (key) DO NOTHING;
