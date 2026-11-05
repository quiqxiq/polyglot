-- 000041_technician_role_merge.up.sql
-- Merges the legacy "technician" role into the canonical "teknisi" role.
--
-- users.role is the source of truth. Login and application startup reconcile
-- Casbin grouping policies from that column, but this migration updates the
-- persisted data immediately so already-issued tokens and currently running
-- instances observe a consistent role until the next login/restart.
--
-- casbin_rule is created lazily by the GORM adapter on application startup, so
-- the Casbin section is guarded for fresh databases where migrations run before
-- the server has ever booted.

UPDATE users
SET role = 'teknisi',
    updated_at = CURRENT_TIMESTAMP
WHERE LOWER(TRIM(role)) = 'technician';

DO $$
BEGIN
    IF to_regclass('public.casbin_rule') IS NOT NULL THEN
        -- Grouping policy: g, user, role.
        UPDATE casbin_rule
        SET v1 = 'teknisi'
        WHERE ptype = 'g' AND v1 = 'technician';

        -- Defensive: a grouping row could theoretically carry the role in v0.
        UPDATE casbin_rule
        SET v0 = 'teknisi'
        WHERE ptype = 'g' AND v0 = 'technician';

        -- Policy rows for the legacy role: p, role, object, action.
        UPDATE casbin_rule
        SET v0 = 'teknisi'
        WHERE ptype = 'p' AND v0 = 'technician';
    ELSE
        RAISE NOTICE 'casbin_rule not found; grouping policies will be reconciled on first start';
    END IF;
END $$;
