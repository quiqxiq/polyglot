-- 000023_add_username_pattern_settings.up.sql
-- Default setting pola penamaan username PPPoE/Hotspot & mode password otomatis

INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.pppoe_username_pattern', '{initials}{digits4}', 'isp_credential',
     'Pola template pembuatan username otomatis: {initials}, {name_slug}, {customer_code}, {digits4}, {digits6}'),
    ('isp.pppoe_username_prefix', '', 'isp_credential',
     'Prefix opsional di awal username (misal: "net-")'),
    ('isp.pppoe_password_mode', 'digits6', 'isp_credential',
     'Mode pembuatan password default: digits6 (6 angka), random8 (8 alfanumerik), phone (no wa)')
ON CONFLICT (key) DO NOTHING;
