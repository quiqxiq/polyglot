-- 000034_router_jobs.down.sql
DELETE FROM system_settings WHERE key IN (
    'isp.router_job_batch_size',
    'isp.router_job_error_budget',
    'isp.router_job_cron',
    'isp.drift_cron'
);

DROP TABLE IF EXISTS router_jobs;
