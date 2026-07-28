CREATE TABLE subscription_status_history (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscription_id         UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    old_status              TEXT,
    new_status              TEXT NOT NULL,
    changed_by_user         UUID REFERENCES users(id) ON DELETE SET NULL,
    changed_by_actor_type   TEXT NOT NULL DEFAULT 'human' CHECK (changed_by_actor_type IN ('human','ai_agent','system_scheduled')),
    reason                  TEXT,
    changed_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);
