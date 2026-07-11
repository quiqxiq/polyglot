-- pgcrypto dibutuhkan untuk gen_random_uuid() di seluruh migration berikutnya.
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE odcs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    olt_device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE RESTRICT,
    location_lat    DOUBLE PRECISION,
    location_lng    DOUBLE PRECISION,
    capacity_ports  INTEGER NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE odps (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    odc_id          UUID REFERENCES odcs(id) ON DELETE SET NULL,
    olt_device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE RESTRICT,
    pon_port        TEXT NOT NULL,
    name            TEXT NOT NULL,
    capacity_ports  INTEGER NOT NULL,
    location_lat    DOUBLE PRECISION,
    location_lng    DOUBLE PRECISION,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (olt_device_id, pon_port)
);

CREATE TABLE ip_pools (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id       UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    name            TEXT NOT NULL,
    cidr            CIDR NOT NULL,
    gateway         INET,
    pool_type       TEXT NOT NULL CHECK (pool_type IN ('pppoe', 'hotspot', 'static')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (device_id, name)
);
