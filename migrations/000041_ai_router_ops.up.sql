-- 000041_ai_router_ops.up.sql
-- PLAN-AI-ROUTER-OPS v2 (M1): restore point, approval HITL, command logs,
-- active device binding, dan persistensi AI chat.
-- Prinsip: setiap mutation AI wajib approval manusia + restore point;
-- AI tidak pernah bisa approve request sendiri.

-- ─── 1. Restore point bundle ───────────────────────────────────────────────
CREATE TABLE restore_points (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL DEFAULT 'tenant-default',
    device_id         UUID REFERENCES devices(id) ON DELETE SET NULL,
    trigger           TEXT NOT NULL,
    scope             JSONB NOT NULL DEFAULT '[]',
    routeros_version  TEXT,
    config_hash       TEXT,
    rollback_script   JSONB,
    manifest          JSONB NOT NULL DEFAULT '{}',
    total_size_bytes  BIGINT NOT NULL DEFAULT 0,
    status            TEXT NOT NULL DEFAULT 'CREATING',
    created_by        TEXT NOT NULL DEFAULT '',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at        TIMESTAMPTZ NOT NULL,
    used_at           TIMESTAMPTZ,
    CONSTRAINT chk_restore_trigger CHECK (trigger IN ('MUTATION','SCHEDULED','MANUAL')),
    CONSTRAINT chk_restore_status CHECK (status IN ('CREATING','AVAILABLE','PARTIAL','USED','EXPIRED','DELETED'))
);
CREATE INDEX idx_restore_device ON restore_points (device_id, created_at DESC);
CREATE INDEX idx_restore_status ON restore_points (status, expires_at);
CREATE INDEX idx_restore_config_hash ON restore_points (device_id, config_hash);
CREATE TRIGGER trg_restore_points_upd BEFORE UPDATE ON restore_points
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE restore_point_artifacts (
    id                BIGSERIAL PRIMARY KEY,
    restore_point_id  TEXT NOT NULL REFERENCES restore_points(id) ON DELETE CASCADE,
    kind              TEXT NOT NULL,
    scope             JSONB NOT NULL DEFAULT '[]',
    storage_path      TEXT NOT NULL,
    sha256            TEXT NOT NULL,
    size_bytes        BIGINT NOT NULL DEFAULT 0,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_restore_artifact_kind CHECK (kind IN ('SCOPED_EXPORT','BINARY_BACKUP','INVERSE_SCRIPT')),
    CONSTRAINT uq_restore_artifact UNIQUE (restore_point_id, kind, storage_path)
);
CREATE INDEX idx_restore_artifacts ON restore_point_artifacts (restore_point_id);

-- ─── 2. Approval HITL ──────────────────────────────────────────────────────
CREATE TABLE approval_requests (
    id                TEXT PRIMARY KEY,
    tenant_id         TEXT NOT NULL DEFAULT 'tenant-default',
    device_id         UUID REFERENCES devices(id) ON DELETE SET NULL,
    session_id        TEXT NOT NULL DEFAULT '',
    actor_type        TEXT NOT NULL,
    actor_id          TEXT NOT NULL DEFAULT '',
    requested_by      TEXT NOT NULL DEFAULT '',
    command           TEXT NOT NULL,
    command_args      JSONB NOT NULL DEFAULT '{}',
    revision          INT NOT NULL DEFAULT 1,
    edited_command    TEXT,
    edited_args       JSONB,
    edited_by         BIGINT REFERENCES users(id) ON DELETE SET NULL,
    edit_count        INT NOT NULL DEFAULT 0,
    diff              JSONB NOT NULL DEFAULT '{}',
    validation        JSONB,
    restore_point_id  TEXT REFERENCES restore_points(id) ON DELETE SET NULL,
    status            TEXT NOT NULL DEFAULT 'PENDING',
    reason            TEXT,
    approved_by       BIGINT REFERENCES users(id) ON DELETE SET NULL,
    decided_by        BIGINT REFERENCES users(id) ON DELETE SET NULL,
    idempotency_key   TEXT,
    execution_nonce   TEXT,
    execution_result  JSONB,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at        TIMESTAMPTZ NOT NULL,
    decided_at        TIMESTAMPTZ,
    executed_at       TIMESTAMPTZ,
    CONSTRAINT chk_approval_status CHECK (status IN ('PENDING','APPROVED','EXECUTING','EXECUTED','FAILED','REJECTED','EXPIRED','CANCELLED')),
    CONSTRAINT chk_approval_actor CHECK (actor_type IN ('AGENT','USER','SYSTEM'))
);
CREATE INDEX idx_approval_pending ON approval_requests (tenant_id, status, created_at DESC);
CREATE INDEX idx_approval_device ON approval_requests (device_id, created_at DESC);
CREATE UNIQUE INDEX uq_approval_idempotency ON approval_requests (tenant_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
CREATE TRIGGER trg_approval_requests_upd BEFORE UPDATE ON approval_requests
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ─── 3. Command logs (semua percobaan, termasuk yang ditolak) ──────────────
CREATE TABLE command_logs (
    id                BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tenant_id         TEXT NOT NULL DEFAULT 'tenant-default',
    device_id         UUID REFERENCES devices(id) ON DELETE SET NULL,
    actor_type        TEXT NOT NULL,
    actor_id          TEXT,
    session_id        TEXT,
    mode              TEXT NOT NULL,
    command           TEXT NOT NULL,
    command_args      JSONB,
    approval_id       TEXT REFERENCES approval_requests(id) ON DELETE SET NULL,
    restore_point_id  TEXT REFERENCES restore_points(id) ON DELETE SET NULL,
    status            TEXT NOT NULL,
    output_summary    TEXT,
    duration_ms       BIGINT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT chk_command_log_mode CHECK (mode IN ('READ','MUTATION')),
    CONSTRAINT chk_command_log_status CHECK (status IN ('SUCCESS','FAILED','DENIED','PENDING_APPROVAL','REJECTED','EXPIRED'))
);
CREATE INDEX idx_command_logs_device_time ON command_logs (device_id, created_at DESC);
CREATE INDEX idx_command_logs_actor ON command_logs (actor_type, actor_id, created_at DESC);

-- ─── 4. Active device binding (fallback/audit; Redis utama) ────────────────
CREATE TABLE active_devices (
    tenant_id  TEXT NOT NULL,
    user_id    TEXT NOT NULL,
    session_id TEXT NOT NULL DEFAULT '',
    device_id  UUID REFERENCES devices(id) ON DELETE CASCADE,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (tenant_id, user_id, session_id)
);

-- ─── 5. AI chat persistence ────────────────────────────────────────────────
CREATE TABLE ai_conversations (
    id              TEXT PRIMARY KEY,
    tenant_id       TEXT NOT NULL DEFAULT 'tenant-default',
    user_id         BIGINT REFERENCES users(id) ON DELETE CASCADE,
    device_id       UUID REFERENCES devices(id) ON DELETE SET NULL,
    model_config_id TEXT,
    title           TEXT,
    status          TEXT NOT NULL DEFAULT 'ACTIVE',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at     TIMESTAMPTZ,
    CONSTRAINT chk_ai_conversation_status CHECK (status IN ('ACTIVE','ARCHIVED'))
);
CREATE INDEX idx_ai_conversations_user ON ai_conversations (tenant_id, user_id, created_at DESC);
CREATE TRIGGER trg_ai_conversations_upd BEFORE UPDATE ON ai_conversations
FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE ai_chat_events (
    id              BIGSERIAL PRIMARY KEY,
    conversation_id TEXT NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    seq             BIGINT NOT NULL,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT uq_ai_chat_event_seq UNIQUE (conversation_id, seq),
    CONSTRAINT chk_ai_event_type CHECK (event_type IN (
        'text','reasoning','tool_call','tool_result','plan_step','terminal',
        'approval','checkpoint','usage','done','error'))
);
CREATE INDEX idx_ai_events_conversation ON ai_chat_events (conversation_id, seq);
