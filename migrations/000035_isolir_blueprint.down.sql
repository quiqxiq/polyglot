-- 000035_isolir_blueprint.down.sql
DELETE FROM system_settings WHERE key IN (
    'isp.isolir_webproxy_enabled',
    'isp.isolir_webproxy_port',
    'isp.isolir_billing_hosts',
    'isp.isolir_error_page_url'
);
