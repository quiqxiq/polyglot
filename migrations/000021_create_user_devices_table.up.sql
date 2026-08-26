-- 000021_create_user_devices_table.up.sql
-- Relasi Many-to-Many antara user dan router (device) untuk data-level scoping / RBAC

CREATE TABLE IF NOT EXISTS user_devices (
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id   UUID NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    assigned_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, device_id)
);

CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices(user_id);
CREATE INDEX IF NOT EXISTS idx_user_devices_device_id ON user_devices(device_id);
