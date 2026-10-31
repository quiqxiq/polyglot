-- 000037_acs.down.sql
DELETE FROM system_settings WHERE key LIKE 'acs.%';

DROP TABLE IF EXISTS acs_presets;
DROP TABLE IF EXISTS acs_devices;
