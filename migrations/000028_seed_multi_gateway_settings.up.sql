-- 000028_seed_multi_gateway_settings.up.sql
-- Konfigurasi multi payment gateway (F4-5): Tripay tetap default; Midtrans &
-- Xendit nonaktif sampai admin mengisi kredensial.
INSERT INTO system_settings (key, value, category, description) VALUES
    ('gw.active', 'TRIPAY', 'isp_gateway', 'Gateway pembayaran aktif default'),
    ('gw.midtrans.enabled', 'false', 'isp_gateway', 'Aktifkan Midtrans'),
    ('gw.midtrans.endpoint', 'https://api.sandbox.midtrans.com', 'isp_gateway', 'Base URL Midtrans (Snap/Core)'),
    ('gw.midtrans.server_key', '', 'isp_gateway', 'Server Key Midtrans (terenkripsi)'),
    ('gw.midtrans.client_key', '', 'isp_gateway', 'Client Key Midtrans'),
    ('gw.midtrans.channel', '', 'isp_gateway', 'Channel default Midtrans (kosong = semua)'),
    ('gw.xendit.enabled', 'false', 'isp_gateway', 'Aktifkan Xendit'),
    ('gw.xendit.endpoint', 'https://api.xendit.co', 'isp_gateway', 'Base URL Xendit'),
    ('gw.xendit.secret_key', '', 'isp_gateway', 'Secret Key Xendit (terenkripsi)'),
    ('gw.xendit.callback_token', '', 'isp_gateway', 'Callback verification token Xendit (terenkripsi)'),
    ('gw.xendit.channel', '', 'isp_gateway', 'Channel default Xendit (kosong = semua)')
ON CONFLICT (key) DO NOTHING;
