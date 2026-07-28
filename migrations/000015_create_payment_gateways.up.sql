CREATE TABLE payment_gateways (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL UNIQUE,
    provider    TEXT NOT NULL CHECK (provider IN ('midtrans','xendit','manual')),
    config_ref  TEXT,
    is_active   BOOLEAN NOT NULL DEFAULT true
);
