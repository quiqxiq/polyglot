CREATE TABLE device_stream_subscriptions (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id           UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    requested_by_user   UUID REFERENCES users(id) ON DELETE SET NULL,
    command_raw         TEXT NOT NULL,
    started_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    ended_at            TIMESTAMPTZ,
    end_reason          TEXT CHECK (end_reason IN ('client_disconnected','cancelled','device_error','completed'))
);
