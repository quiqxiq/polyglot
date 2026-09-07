-- 000024_optimize_ping_metrics_timescale.up.sql
-- Optimasi performa dan kompresi TimescaleDB untuk device_ping_metrics

-- 1. Sesuaikan interval chunk menjadi 1 hari (bukan default 7 hari) agar working chunk ramah memori RAM
SELECT set_chunk_time_interval('device_ping_metrics', INTERVAL '1 day');

-- 2. Hapus foreign key constraint jika ada, untuk menghindari referential lock checking per row insert
ALTER TABLE IF EXISTS device_ping_metrics DROP CONSTRAINT IF EXISTS device_ping_metrics_device_id_fkey;

-- 3. Aktifkan TimescaleDB Columnar Compression
ALTER TABLE device_ping_metrics SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'device_id, target',
    timescaledb.compress_orderby = 'recorded_at DESC'
);

-- 4. Pasang kebijakan kompresi otomatis untuk chunk data lebih dari 2 jam
SELECT add_compression_policy('device_ping_metrics', INTERVAL '2 hours', if_not_exists => true);

-- 5. Buat Materialized Continuous Aggregate View per 1 menit
CREATE MATERIALIZED VIEW IF NOT EXISTS device_ping_metrics_1m
WITH (timescaledb.continuous) AS
SELECT 
    time_bucket('1 minute', recorded_at) AS bucket_time,
    device_id,
    target,
    AVG(NULLIF(rtt_ms, 0)) AS avg_rtt_ms,
    MIN(NULLIF(rtt_ms, 0)) AS min_rtt_ms,
    MAX(NULLIF(rtt_ms, 0)) AS max_rtt_ms,
    SUM(sent) AS sent,
    SUM(received) AS received,
    COUNT(*) AS sample_count
FROM device_ping_metrics
GROUP BY bucket_time, device_id, target
WITH NO DATA;

-- Aktifkan Real-Time Aggregation agar data terbaru yang belum termaterialisasi tetap muncul dalam query
ALTER MATERIALIZED VIEW device_ping_metrics_1m SET (timescaledb.materialized_only = false);

-- 6. Jadwalkan refresh otomatis untuk Continuous Aggregate View
SELECT add_continuous_aggregate_policy('device_ping_metrics_1m',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute',
    if_not_exists => true);
