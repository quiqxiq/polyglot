-- 000015_add_burst_to_service_plans.down.sql

ALTER TABLE service_plans
    DROP COLUMN IF EXISTS burst_download_kbps,
    DROP COLUMN IF EXISTS burst_upload_kbps,
    DROP COLUMN IF EXISTS burst_threshold_kbps,
    DROP COLUMN IF EXISTS burst_time_seconds;
