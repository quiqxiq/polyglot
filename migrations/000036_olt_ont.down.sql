-- 000036_olt_ont.down.sql
DELETE FROM notification_templates WHERE id IN ('nt-ont-offline','nt-ont-online','nt-ont-low-rx');

DELETE FROM system_settings WHERE key LIKE 'olt.%' OR key LIKE 'alert.ont%';

DROP TABLE IF EXISTS ont_metric_samples;
DROP TABLE IF EXISTS ont_events;
DROP TABLE IF EXISTS onts;

ALTER TABLE registrations
    DROP COLUMN IF EXISTS onu_index,
    DROP COLUMN IF EXISTS pon_index,
    DROP COLUMN IF EXISTS olt_device_id;
