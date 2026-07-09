-- 000001: initial schema — devices + credentials
--
-- Two tables, deliberately separated per Polyglot-Architecture.md §2.4
-- ("AI tidak pernah menyentuh kredensial mentah") and §7.2:
--   devices     — plaintext connection params + non-sensitive vendor config
--   credentials — encrypted blob (AES-GCM), 1:1 to device, never exposed
--                 above the repository/vault layer
--
-- The repository layer merges a devices row (plaintext extra) with a
-- decrypted credentials blob into a single device.Target for the driver,
-- so drivers see one uniform Extra map regardless of where each field
-- physically lives.

CREATE TABLE devices (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL,
    -- vendor: hardware vendor (mikrotik, cisco, zte, huawei, genieacs, ...).
    -- driver_type: Go driver package to instantiate (mikrotik, cisco,
    -- zteolt, huaweiolt, genieacs, genericssh, ...). These are distinct —
    -- a GenieACS-managed CPE may have vendor "zte" but driver_type
    -- "genieacs", because the CPE is reached through the ACS, not directly.
    vendor           TEXT NOT NULL,
    driver_type      TEXT NOT NULL,
    host             TEXT NOT NULL,
    -- port 0 means "use the driver's default" (e.g. 7557 for genieacs,
    -- 8728 for mikrotik API). The driver resolves it at connect time.
    port             INTEGER NOT NULL DEFAULT 0,
    timeout_ms       INTEGER NOT NULL DEFAULT 30000,
    poll_interval_ms INTEGER NOT NULL DEFAULT 30000,
    -- extra: non-sensitive vendor-specific params as JSONB. Examples:
    --   genieacs: {"device_id":"...","use_tls":"true","fault_channel":"default"}
    --   zteolt:   {"snmp_port":"161"}
    -- Sensitive values (password, community string, api_key) do NOT go
    -- here — they live in the credentials blob.
    extra            JSONB NOT NULL DEFAULT '{}',
    tags             TEXT[] NOT NULL DEFAULT '{}',
    enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_devices_driver_type ON devices (driver_type);
CREATE INDEX idx_devices_vendor      ON devices (vendor);
CREATE INDEX idx_devices_tags        ON devices USING GIN (tags);

CREATE TABLE credentials (
    -- One credential row per device (1:1). PK is device_id itself — no
    -- separate surrogate key needed, and it enforces the cardinality.
    device_id  UUID PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
    -- ciphertext: AES-GCM encrypted JSON blob. Shape is vendor-dependent
    -- (e.g. {"username":"...","password":"..."} for SSH, {"community":"..."}
    -- for SNMP, {"api_key":"..."} for genieacs x-api-key). Storing a blob
    -- rather than fixed username/password columns avoids altering this
    -- table every time a vendor needs a new credential field.
    ciphertext BYTEA NOT NULL,
    -- nonce: AES-GCM 12-byte nonce, unique per encryption (not reused).
    -- Stored alongside ciphertext so the vault can decrypt without a
    -- second lookup.
    nonce      BYTEA NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- updated_at auto-maintenance: a trigger keeps it honest even for raw SQL
-- UPDATEs that bypass GORM's autoUpdateTime hook.
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_devices_updated_at
    BEFORE UPDATE ON devices
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TRIGGER trg_credentials_updated_at
    BEFORE UPDATE ON credentials
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
