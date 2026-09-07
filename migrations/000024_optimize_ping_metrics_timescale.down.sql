-- 000024_optimize_ping_metrics_timescale.down.sql
-- Rollback optimasi dan kompresi TimescaleDB untuk device_ping_metrics

-- 1. Hapus policy dan view Continuous Aggregate
SELECT remove_continuous_aggregate_policy('device_ping_metrics_1m', if_exists => true);
DROP MATERIALIZED VIEW IF EXISTS device_ping_metrics_1m;

-- 2. Hapus compression policy dan nonaktifkan kompresi
SELECT remove_compression_policy('device_ping_metrics', if_exists => true);
ALTER TABLE IF EXISTS device_ping_metrics SET (timescaledb.compress = false);

-- 3. Kembalikan chunk interval ke default 7 hari
SELECT set_chunk_time_interval('device_ping_metrics', INTERVAL '7 days');
