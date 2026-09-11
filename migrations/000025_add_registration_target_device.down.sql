-- 000025_add_registration_target_device.down.sql
DROP INDEX IF EXISTS idx_registrations_target_device;
ALTER TABLE registrations DROP COLUMN IF EXISTS target_device_id;
