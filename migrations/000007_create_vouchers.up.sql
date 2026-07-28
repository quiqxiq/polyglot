CREATE TABLE voucher_batches (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    plan_id             UUID NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,
    quantity_generated  INTEGER NOT NULL,
    price_per_voucher   NUMERIC(14,2) NOT NULL,
    generated_by        UUID REFERENCES users(id) ON DELETE SET NULL,
    generated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- used_by_subscription_id sengaja UUID polos (bukan REFERENCES) di sini —
-- tabel subscriptions belum ada sampai migration 000009. FK-nya ditutup
-- di 000009_create_subscriptions.up.sql lewat ALTER TABLE, setelah
-- subscriptions benar-benar ada. Lihat DATABASE-SCHEMA.md §4.4 untuk
-- penjelasan lengkap kenapa pola ini dipakai (bukan kelalaian).
CREATE TABLE hotspot_vouchers (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_id                    UUID REFERENCES voucher_batches(id) ON DELETE SET NULL,
    plan_id                     UUID NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,
    code                        TEXT NOT NULL UNIQUE,
    status                      TEXT NOT NULL DEFAULT 'unused' CHECK (status IN ('unused','active','expired','used')),
    used_by_subscription_id     UUID,
    used_by_mac                 MACADDR,
    activated_at                TIMESTAMPTZ,
    expires_at                  TIMESTAMPTZ,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);
