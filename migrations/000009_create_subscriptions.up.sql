CREATE TABLE subscriptions (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    customer_id                 UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    plan_id                     UUID NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,
    service_type                TEXT NOT NULL CHECK (service_type IN ('pppoe','hotspot','static_ip','dhcp')),
    status                      TEXT NOT NULL DEFAULT 'pending_install'
                                    CHECK (status IN ('pending_install','active','suspended','terminated')),
    device_id                   UUID NOT NULL REFERENCES devices(id) ON DELETE RESTRICT,
    odp_id                      UUID REFERENCES odps(id) ON DELETE SET NULL,
    odp_port                    TEXT,
    onu_serial_number           TEXT,
    ip_pool_id                  UUID REFERENCES ip_pools(id) ON DELETE SET NULL,
    pppoe_username              TEXT,
    pppoe_password_encrypted    TEXT,
    static_ip                   INET,
    mac_address                 MACADDR,
    installed_at                TIMESTAMPTZ,
    activated_at                TIMESTAMPTZ,
    suspended_at                TIMESTAMPTZ,
    terminated_at                TIMESTAMPTZ,
    suspension_reason            TEXT,
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (device_id, pppoe_username)
);

-- Menutup FK yang ditunda dari 000007_create_vouchers — subscriptions
-- sekarang sudah ada.
ALTER TABLE hotspot_vouchers
    ADD CONSTRAINT fk_voucher_subscription
    FOREIGN KEY (used_by_subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL;
