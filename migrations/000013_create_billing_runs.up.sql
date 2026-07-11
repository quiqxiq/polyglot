CREATE TABLE billing_runs (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    period                  DATE NOT NULL,
    generated_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    generated_by            UUID REFERENCES users(id) ON DELETE SET NULL,
    total_invoices_created  INTEGER NOT NULL DEFAULT 0,
    status                  TEXT NOT NULL DEFAULT 'running' CHECK (status IN ('running','completed','failed')),
    UNIQUE (period)
);
