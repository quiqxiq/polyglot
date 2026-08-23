-- 000018_isp_settings_and_provision_status.down.sql

ALTER TABLE subscriptions DROP COLUMN IF EXISTS provision_status;

DELETE FROM system_settings WHERE key LIKE 'isp.%';
