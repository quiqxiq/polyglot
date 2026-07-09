-- Reverse order of up.sql: drop triggers/functions first, then tables
-- (credentials before devices due to FK dependency).
DROP TRIGGER IF EXISTS trg_credentials_updated_at ON credentials;
DROP TRIGGER IF EXISTS trg_devices_updated_at ON devices;
DROP FUNCTION IF EXISTS set_updated_at();
DROP TABLE IF EXISTS credentials;
DROP TABLE IF EXISTS devices;
