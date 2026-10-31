-- 000035_isolir_blueprint.up.sql
-- Fase 4 D12: blueprint sistem isolir §5.4 analisis — webproxy internal,
-- whitelist server billing, dan error.html mode self/cloud.

INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.isolir_webproxy_enabled', 'true', 'isp_isolation',
     'Aktifkan webproxy MikroTik untuk halaman isolir lokal (mode self)'),
    ('isp.isolir_webproxy_port', '3125', 'isp_isolation',
     'Port webproxy isolir di MikroTik (default 3125)'),
    ('isp.isolir_billing_hosts', '', 'isp_isolation',
     'Daftar host/IP server billing yang diizinkan saat isolir (dipisah koma)'),
    ('isp.isolir_error_page_url', '', 'isp_isolation',
     'URL error.html untuk /tool fetch ke flash router (mode self); kosong = mode cloud')
ON CONFLICT (key) DO NOTHING;
