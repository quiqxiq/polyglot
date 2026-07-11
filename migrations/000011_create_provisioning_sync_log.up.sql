-- command_audit_log_id sengaja UUID polos (bukan REFERENCES) — tabel
-- command_audit_log belum ada sampai migration 000017. FK-nya ditutup di
-- situ lewat ALTER TABLE. Lihat DATABASE-SCHEMA.md §6.3/§7 untuk konteks
-- lengkap kenapa tabel ini jadi jantung sinkronisasi ke MikroTik.
CREATE TABLE provisioning_sync_log (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id         UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    target_type             TEXT NOT NULL CHECK (target_type IN
                                ('mikrotik_ppp_secret','mikrotik_hotspot_user','mikrotik_address_list','freeradius','genieacs_tr069')),
    device_id               UUID REFERENCES devices(id) ON DELETE SET NULL,
    action                  TEXT NOT NULL CHECK (action IN ('create','update','disable','enable','delete','change_profile')),
    status                  TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','success','failed')),
    requested_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at            TIMESTAMPTZ,
    error_message           TEXT,
    command_audit_log_id    UUID,
    external_reference      TEXT
);
