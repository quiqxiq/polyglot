-- 000023_add_username_pattern_settings.down.sql
DELETE FROM system_settings WHERE key IN (
    'isp.pppoe_username_pattern',
    'isp.pppoe_username_prefix',
    'isp.pppoe_password_mode'
);
