CREATE TABLE plan_router_profiles (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id                 UUID NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    device_id               UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    mikrotik_profile_name   TEXT NOT NULL,
    sync_status             TEXT NOT NULL DEFAULT 'pending' CHECK (sync_status IN ('pending','synced','error')),
    last_synced_at          TIMESTAMPTZ,
    sync_error_message      TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (plan_id, device_id)
);
