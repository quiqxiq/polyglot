-- 000025_add_registration_target_device.up.sql
-- Router BRAS yang dipilih teknisi saat pemasangan (F2-7).
ALTER TABLE registrations
    ADD COLUMN IF NOT EXISTS target_device_id UUID REFERENCES devices(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_registrations_target_device ON registrations(target_device_id);
