-- 000038_billing_ops_v3.up.sql
-- W0 PLAN-BILLING-OPS-V3: mode billing (prabayar/pascabayar), tanggal tutup
-- buku, tampilan "Belum Bayar H-N", PPN default mitra, offset reminder
-- pre/post jatuh tempo, dan nomor invoice berurutan.
-- CATATAN KEPUTUSAN: tidak ada denda keterlambatan dan tidak ada biaya
-- tambahan/extra fee pada tagihan (otomatis maupun manual).

-- PPN default mitra (fallback bila plan.tax_percent = 0); invoice menyimpan
-- tarif efektif yang dipakai untuk audit & laporan pajak.
INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.billing_mode',              'PASCABAYAR', 'isp_billing',
     'Mode billing: PRABAYAR (bayar dulu baru pakai) atau PASCABAYAR (pakai dulu baru bayar)'),
    ('isp.closing_date',              '1',          'isp_billing',
     'Tanggal tutup buku/siklus tagihan (1-28); default anchor langganan baru'),
    ('isp.unpaid_days_display',       '7',          'isp_billing',
     'Ambang tampilan status Belum Bayar (H-N) di daftar pelanggan/tagihan; hanya tampilan'),
    ('isp.prepay_invoice_lead_days',  '7',          'isp_billing',
     'Mode PRABAYAR: berapa hari sebelum siklus tagihan diterbitkan & diingatkan'),
    ('isp.tax_percent',               '0',          'isp_billing',
     'PPN default mitra (persen) bila tarif pajak paket kosong'),
    ('isp.reminder_pre_offsets',      '7,3,1,0',    'isp_billing',
     'Offset pengingat sebelum jatuh tempo (hari), dipisah koma'),
    ('isp.reminder_post_offsets',     '1,3,7',      'isp_billing',
     'Offset pengingat setelah jatuh tempo (hari), dipisahkan koma'),
    ('isp.gateway_sync_stale_minutes','15',         'isp_gateway',
     'Transaksi gateway PENDING lebih dari N menit akan disinkronkan/di-expire')
ON CONFLICT (key) DO NOTHING;

ALTER TABLE invoices
    ADD COLUMN IF NOT EXISTS tax_percent NUMERIC(5,2) NOT NULL DEFAULT 0.00;

CREATE SEQUENCE IF NOT EXISTS invoice_no_seq;

INSERT INTO notification_templates (id, template_key, name, content, variables_json) VALUES
    ('nt-invoice-created', 'INVOICE_CREATED',
     'Tagihan Baru Terbit',
     'Yth {{customer_name}}, tagihan periode {{period}} sebesar Rp{{total}} telah terbit. Jatuh tempo {{due_date}}. Kode bayar: {{payment_code}}.',
     '["customer_name","period","total","due_date","payment_code"]'),
    ('nt-bill-overdue', 'BILL_OVERDUE',
     'Tagihan Melewati Jatuh Tempo',
     'Yth {{customer_name}}, tagihan periode {{period}} sebesar Rp{{total}} telah melewati jatuh tempo {{due_date}}. Mohon segera lakukan pembayaran. Kode bayar: {{payment_code}}.',
     '["customer_name","period","total","due_date","payment_code"]')
ON CONFLICT (id) DO NOTHING;
