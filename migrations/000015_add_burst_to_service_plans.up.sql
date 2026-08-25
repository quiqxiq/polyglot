-- 000015_add_burst_to_service_plans.up.sql
--
-- Dukungan burst limit untuk master paket. Kolom opsional: terisi semua
-- atau kosong semua (tanpa burst). Nilai diterjemahkan menjadi string
-- rate-limit RouterOS saat provisioning:
--   "10M/10M 20M/20M 15M/15M 16s"
-- (rx-rate/tx-rate rx-burst-rate/tx-burst-rate rx-burst-threshold/tx-burst-threshold rx-burst-time/tx-burst-time)
-- Nama kolom selaras dengan DATABASE-SCHEMA-ISP.md.

ALTER TABLE service_plans
    ADD COLUMN burst_download_kbps  INT,
    ADD COLUMN burst_upload_kbps    INT,
    ADD COLUMN burst_threshold_kbps INT,
    ADD COLUMN burst_time_seconds   INT;
