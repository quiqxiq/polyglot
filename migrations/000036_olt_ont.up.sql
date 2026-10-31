-- 000036_olt_ont.up.sql
-- Fase A (PLAN-ACS-OLT-HIOSO.md): inventori OLT EPON HiOSO + ONT/ONU,
-- event timeline, sampel metrik optik, integrasi PSB (OLT/PON/ONU), dan
-- seed konfigurasi polling/alert.

-- PSB: teknisi memilih OLT/PON saat instalasi; MAC/SN sudah ada sejak 000032.
ALTER TABLE registrations
    ADD COLUMN IF NOT EXISTS olt_device_id UUID REFERENCES devices(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS pon_index INT,
    ADD COLUMN IF NOT EXISTS onu_index INT;

-- §Inventori ONT/ONU per OLT (identity = MAC LLID, unik per OLT).
CREATE TABLE onts (
    id                 TEXT PRIMARY KEY,
    tenant_id          TEXT NOT NULL DEFAULT 'tenant-default',
    olt_device_id      UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    pon_index          INT NOT NULL,
    onu_index          INT NOT NULL,
    llid               INT,
    name               VARCHAR(100),
    mac_address        VARCHAR(20) NOT NULL,
    serial_number      VARCHAR(50),
    model              VARCHAR(50),
    firmware           VARCHAR(50),
    state              VARCHAR(20) NOT NULL DEFAULT 'UNKNOWN',
    admin_state        VARCHAR(10) NOT NULL DEFAULT 'ENABLE',
    rx_power_dbm       NUMERIC(6,2),
    tx_power_dbm       NUMERIC(6,2),
    distance_m         INT,
    temperature_c      NUMERIC(5,1),
    missed_polls       INT NOT NULL DEFAULT 0,
    offline_strikes    INT NOT NULL DEFAULT 0,
    rx_na_strikes      INT NOT NULL DEFAULT 0,
    last_seen_at       TIMESTAMPTZ,
    last_polled_at     TIMESTAMPTZ,
    customer_id        TEXT REFERENCES customers(id) ON DELETE SET NULL,
    subscription_id    TEXT REFERENCES subscriptions(id) ON DELETE SET NULL,
    pppoe_username     VARCHAR(100),
    acs_device_id      VARCHAR(120),
    acs_matched_by     VARCHAR(20),
    acs_last_inform_at TIMESTAMPTZ,
    acs_telemetry      JSONB,
    wlan_slots         INT NOT NULL DEFAULT 0,
    source             VARCHAR(10) NOT NULL DEFAULT 'SNMP',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_ont_state CHECK (state IN ('ONLINE','OFFLINE','UNKNOWN','REMOVED')),
    CONSTRAINT chk_ont_admin_state CHECK (admin_state IN ('ENABLE','DISABLE')),
    CONSTRAINT uq_onts_olt_mac UNIQUE (olt_device_id, mac_address)
);

-- Satu posisi PON/ONU hanya boleh dipakai satu ONT aktif; baris REMOVED
-- (ONU diganti/dicabut) tetap disimpan sebagai histori.
CREATE UNIQUE INDEX uq_onts_olt_pos_active ON onts(olt_device_id, pon_index, onu_index)
    WHERE state <> 'REMOVED';

CREATE INDEX idx_onts_state ON onts(state);
CREATE INDEX idx_onts_subscription ON onts(subscription_id);
CREATE INDEX idx_onts_customer ON onts(customer_id);
CREATE INDEX idx_onts_acs ON onts(acs_device_id);
CREATE TRIGGER trg_onts_upd BEFORE UPDATE ON onts
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- Timeline per ONT (discovered/online/offline/los/rename/reboot/acs/link).
CREATE TABLE ont_events (
    id          TEXT PRIMARY KEY,
    tenant_id   TEXT NOT NULL DEFAULT 'tenant-default',
    ont_id      TEXT NOT NULL REFERENCES onts(id) ON DELETE CASCADE,
    event_type  VARCHAR(40) NOT NULL,
    old_state   VARCHAR(20),
    new_state   VARCHAR(20),
    message     TEXT,
    metadata    JSONB,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_ont_events_ont ON ont_events(ont_id, created_at DESC);
CREATE INDEX idx_ont_events_type ON ont_events(event_type, created_at DESC);

-- Sampel metrik optik (retensi harian via scheduler).
CREATE TABLE ont_metric_samples (
    id             TEXT PRIMARY KEY,
    ont_id         TEXT NOT NULL REFERENCES onts(id) ON DELETE CASCADE,
    sampled_at     TIMESTAMPTZ NOT NULL,
    rx_power_dbm   NUMERIC(6,2),
    tx_power_dbm   NUMERIC(6,2),
    temperature_c  NUMERIC(5,1),
    distance_m     INT,
    online         BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE INDEX idx_ont_samples_ont ON ont_metric_samples(ont_id, sampled_at DESC);

-- Konfigurasi polling OLT + alert ONT.
INSERT INTO system_settings (key, value, category, description) VALUES
    ('olt.snmp_timeout_seconds', '5', 'olt', 'Timeout SNMP per request (detik); minimal 5 untuk OLT HiOSO'),
    ('olt.snmp_retries', '2', 'olt', 'Jumlah retry SNMP per request'),
    ('olt.roster_miss_threshold', '12', 'olt', 'Jumlah poll berturut ONT tidak terlihat sebelum dianggap hilang'),
    ('olt.offline_strikes', '2', 'olt', 'Jumlah poll offline berturut sebelum status ONLINE→OFFLINE'),
    ('olt.rx_na_strikes', '3', 'olt', 'Jumlah poll Rx "na" berturut sebelum Rx dianggap tidak tersedia'),
    ('olt.poll_interval_seconds', '300', 'olt', 'Interval sync inventori ONT per OLT (detik)'),
    ('olt.rx_poll_interval_seconds', '900', 'olt', 'Interval polling Rx optik (detik)'),
    ('olt.sample_retention_days', '30', 'olt', 'Retensi sampel metrik optik ONT (hari)'),
    ('olt.sync_cron', '@every 5m', 'olt', 'Jadwal sync OLT otomatis'),
    ('olt.rx_warning_dbm', '-27', 'olt', 'Ambang Rx warning (dBm)'),
    ('olt.rx_critical_dbm', '-30', 'olt', 'Ambang Rx critical/LOS (dBm)'),
    ('alert.ont_cooldown_minutes', '60', 'olt_alert', 'Cooldown notifikasi ONT per perangkat (menit)'),
    ('alert.ont_customer_notify', 'true', 'olt_alert', 'Kirim notifikasi WA ke pelanggan yang tertaut ONT')
ON CONFLICT (key) DO NOTHING;

INSERT INTO notification_templates (id, template_key, name, content, variables_json) VALUES
    ('nt-ont-offline', 'ONT_OFFLINE',
     'Koneksi ONT Terputus',
     'Halo {{customer_name}}, koneksi internet di alamat {{address}} terdeteksi terputus (LOS/offline). Tim teknis kami sedang memeriksa. Jika sudah normal, abaikan pesan ini.',
     '["customer_name","address"]'),
    ('nt-ont-online', 'ONT_ONLINE',
     'Koneksi ONT Pulih',
     'Halo {{customer_name}}, koneksi internet Anda sudah kembali normal. Terima kasih atas kesabarannya.',
     '["customer_name"]'),
    ('nt-ont-low-rx', 'ONT_LOW_RX',
     'Sinyal Optik Menurun',
     'Halo {{customer_name}}, kualitas sinyal optik perangkat Anda menurun ({{rx_power}} dBm). Tim teknis akan melakukan pengecekan.',
     '["customer_name","rx_power"]')
ON CONFLICT (id) DO NOTHING;
