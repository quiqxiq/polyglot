-- TimescaleDB is required for production ping history.
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 2. Buat tabel device_ping_metrics dengan seluruh field telemetri MikroTik
CREATE TABLE IF NOT EXISTS device_ping_metrics (
    recorded_at    TIMESTAMPTZ NOT NULL,
    device_id      UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    target         VARCHAR(255) NOT NULL,
    seq            INTEGER NOT NULL DEFAULT 0,
    size           INTEGER NOT NULL DEFAULT 56,
    ttl            INTEGER NOT NULL DEFAULT 0,
    rtt_ms         REAL NOT NULL DEFAULT 0,
    status         VARCHAR(32) NOT NULL DEFAULT 'connected',
    sent           INTEGER NOT NULL DEFAULT 1,
    received       INTEGER NOT NULL DEFAULT 1,
    packet_loss    SMALLINT NOT NULL DEFAULT 0,
    min_rtt_ms     REAL,
    avg_rtt_ms     REAL,
    max_rtt_ms     REAL
);

-- 3. Inisialisasi Hypertable TimescaleDB
SELECT create_hypertable('device_ping_metrics', 'recorded_at', if_not_exists => TRUE);

-- 4. Indeks komposit untuk optimasi query filter rentang waktu per perangkat
CREATE INDEX IF NOT EXISTS idx_device_ping_metrics_device_time 
    ON device_ping_metrics (device_id, recorded_at DESC);
