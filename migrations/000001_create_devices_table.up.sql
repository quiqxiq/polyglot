CREATE TABLE devices (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name          TEXT NOT NULL,
    vendor        TEXT NOT NULL,
    driver_type   TEXT NOT NULL,
    host          TEXT NOT NULL,
    port          INTEGER NOT NULL,
    timeout_ms    INTEGER NOT NULL DEFAULT 30000,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
