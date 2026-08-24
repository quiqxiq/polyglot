-- 000019_portal_otp_wa_and_settings.down.sql

DELETE FROM system_settings WHERE key LIKE 'gw.tripay.%'
    OR key LIKE 'isp.portal_%' OR key = 'isp.otp_ttl_minutes'
    OR key = 'isp.otp_max_attempts' OR key = 'isp.wa_send_max_retry';

ALTER TABLE wa_notifications DROP COLUMN IF EXISTS attempts;
DROP TABLE IF EXISTS portal_otps;
