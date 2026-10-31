-- 000037_acs.up.sql
-- Fase D (PLAN-ACS-OLT-HIOSO.md): katalog device GenieACS (matching ACS↔ONT)
-- dan preset bootstrap ACS URL + periodic inform.

CREATE TABLE acs_devices (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL DEFAULT 'tenant-default',
    acs_device_id   VARCHAR(120) NOT NULL UNIQUE,
    oui             VARCHAR(20),
    product_class   VARCHAR(80),
    serial_number   VARCHAR(80),
    software_version VARCHAR(80),
    last_inform_at  TIMESTAMPTZ,
    tags            TEXT[] NOT NULL DEFAULT '{}',
    mac_candidates  TEXT[] NOT NULL DEFAULT '{}',
    raw             JSONB,
    matched_ont_id  TEXT REFERENCES onts(id) ON DELETE SET NULL,
    match_source    VARCHAR(20),
    first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_seen_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_acs_devices_serial ON acs_devices(serial_number);
CREATE INDEX idx_acs_devices_matched ON acs_devices(matched_ont_id);
CREATE INDEX idx_acs_devices_inform ON acs_devices(last_inform_at DESC);
CREATE TRIGGER trg_acs_devices_upd BEFORE UPDATE ON acs_devices
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE acs_presets (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL DEFAULT 'tenant-default',
    name            VARCHAR(100) NOT NULL,
    precondition    JSONB,
    configurations  JSONB NOT NULL DEFAULT '[]'::jsonb,
    enabled         BOOLEAN NOT NULL DEFAULT TRUE,
    last_synced_at  TIMESTAMPTZ,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_acs_presets_tenant_name UNIQUE (tenant_id, name)
);
CREATE TRIGGER trg_acs_presets_upd BEFORE UPDATE ON acs_presets
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

INSERT INTO system_settings (key, value, category, description) VALUES
    ('acs.enabled', 'false', 'acs', 'Aktifkan integrasi GenieACS TR-069'),
    ('acs.server_device_id', '', 'acs', 'Devices.id baris server GenieACS (driver_type genieacs, tag acs)'),
    ('acs.periodic_inform_interval', '300', 'acs', 'Periodic Inform Interval ONT (detik)'),
    ('acs.preset_name', 'polyglot-bootstrap', 'acs', 'Nama preset bootstrap ACS di GenieACS'),
    ('acs.param_catalog', '{}', 'acs', 'Override katalog parameter telemetri per ProductClass (JSON)'),
    ('acs.acs_url', '', 'acs', 'ACS URL yang dipush ke ONT (wajib dijangkau ONT, mis. http://acs.example:7547)'),
    ('acs.telemetry_cron', '@every 15m', 'acs', 'Jadwal polling telemetri ACS'),
    ('acs.match_cron', '@every 15m', 'acs', 'Jadwal pencocokan device ACS dengan inventori ONT')
ON CONFLICT (key) DO NOTHING;
