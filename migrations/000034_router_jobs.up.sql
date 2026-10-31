-- 000034_router_jobs.up.sql
-- Fase 4 (planning v2 §3.5): antrean job asynchronous per perangkat router
-- dengan retry exponential backoff, error budget per batch, dan status DEAD.
-- Menggantikan queue client-side JS Global SmartApp.

CREATE TABLE router_jobs (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL DEFAULT 'tenant-default',
    device_id       UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    job_type        VARCHAR(50) NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    idempotency_key VARCHAR(100) UNIQUE NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'QUEUED',
    attempts        INT NOT NULL DEFAULT 0,
    max_attempts    INT NOT NULL DEFAULT 5,
    next_run_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_router_job_status CHECK (status IN ('QUEUED','RUNNING','SUCCEEDED','FAILED','RETRY_WAIT','DEAD'))
);

CREATE INDEX idx_rtr_jobs_poll ON router_jobs(status, next_run_at)
    WHERE status IN ('QUEUED', 'RETRY_WAIT');
CREATE INDEX idx_rtr_jobs_device ON router_jobs(device_id);
CREATE INDEX idx_rtr_jobs_status ON router_jobs(status);

-- Parameter worker (dibaca usecase; nilai default sesuai planning §1.13).
INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.router_job_batch_size', '20', 'isp_billing',
     'Jumlah job router per batch/step (default 20)'),
    ('isp.router_job_error_budget', '5', 'isp_billing',
     'Jumlah error beruntun sebelum batch dihentikan (default 5)'),
    ('isp.router_job_cron', '@every 1m', 'isp_billing',
     'Jadwal worker antrean job router'),
    ('isp.drift_cron', '0 3 * * *', 'isp_billing',
     'Jadwal drift reconciliation router vs database')
ON CONFLICT (key) DO NOTHING;
