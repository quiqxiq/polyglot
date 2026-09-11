-- 000028_seed_multi_gateway_settings.down.sql
DELETE FROM system_settings WHERE key IN (
    'gw.active',
    'gw.midtrans.enabled', 'gw.midtrans.endpoint', 'gw.midtrans.server_key',
    'gw.midtrans.client_key', 'gw.midtrans.channel',
    'gw.xendit.enabled', 'gw.xendit.endpoint', 'gw.xendit.secret_key',
    'gw.xendit.callback_token', 'gw.xendit.channel'
);
