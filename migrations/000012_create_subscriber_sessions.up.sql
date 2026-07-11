-- BUKAN "sessions" — nama sengaja dibedakan dari konsep session driver
-- (internal/domain/session, lihat ADR 0002/0003) yang artinya koneksi KE
-- device, bukan koneksi PELANGGAN ke internet. Lihat DATABASE-SCHEMA.md
-- §6.4 untuk penjelasan lengkap dan rekomendasi pengisian lewat streaming
-- log (bukan polling /ppp active).
CREATE TABLE subscriber_sessions (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id         UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    device_id               UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    external_session_id     TEXT,
    framed_ip               INET,
    mac_address             MACADDR,
    started_at              TIMESTAMPTZ NOT NULL,
    stopped_at              TIMESTAMPTZ,
    bytes_in                BIGINT,
    bytes_out               BIGINT,
    terminate_cause         TEXT
);
