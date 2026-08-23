-- 000014_rebuild_service_plans.up.sql
-- Fase 2 (DATABASE-SCHEMA-ISP.md): tabel plans (migrasi 000006) diganti
-- service_plans dengan parameter MikroTik lengkap. Development stage —
-- tanpa migrasi data lama (keputusan owner): DROP + CREATE deterministik.

DROP TABLE IF EXISTS plans;

CREATE TABLE service_plans (
    id                      TEXT PRIMARY KEY,
    tenant_id               TEXT NOT NULL DEFAULT 'tenant-default',
    name                    VARCHAR(100) NOT NULL,
    service_type            VARCHAR(20)  NOT NULL,               -- 'PPPOE', 'HOTSPOT', 'DEDICATED'
    bandwidth_download_kbps INT          NOT NULL,
    bandwidth_upload_kbps   INT          NOT NULL,
    burst_download_kbps     INT,
    burst_upload_kbps       INT,
    burst_threshold_kbps    INT,
    burst_time_seconds      INT,
    price                   NUMERIC(15,2) NOT NULL,
    selling_price           NUMERIC(15,2),
    installation_fee        NUMERIC(15,2) DEFAULT 0.00,
    tax_percent             NUMERIC(5,2)  DEFAULT 0.00,
    validity                VARCHAR(20)   DEFAULT '30d',
    validity_mode           VARCHAR(20)   DEFAULT 'CALENDAR',
    simultaneous_use        INT           DEFAULT 1,
    ip_pool_name            VARCHAR(50),
    parent_queue            VARCHAR(50)   DEFAULT 'none',
    address_list            VARCHAR(50),
    shared_users            INT           DEFAULT 1,
    expire_mode             VARCHAR(10)   DEFAULT 'ntf',
    lock_user               BOOLEAN       DEFAULT FALSE,
    lock_server             BOOLEAN       DEFAULT FALSE,
    limit_uptime            VARCHAR(20),
    limit_bytes             VARCHAR(20),
    is_active               BOOLEAN       NOT NULL DEFAULT TRUE,
    description             TEXT,
    created_at              TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ   NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_service_plans_tenant_name UNIQUE (tenant_id, name)
);

CREATE INDEX idx_service_plans_type   ON service_plans(service_type);
CREATE INDEX idx_service_plans_active ON service_plans(is_active);

CREATE TRIGGER trg_service_plans_upd BEFORE UPDATE ON service_plans
FOR EACH ROW EXECUTE FUNCTION set_updated_at();
