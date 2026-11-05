-- 000041_technician_role_merge.down.sql
-- Intentionally a no-op.
--
-- The up migration is a lossy rename: after it runs, users.role = 'teknisi'
-- contains both accounts that were already teknisi and accounts migrated from
-- the legacy "technician" role. There is no audit column to distinguish them,
-- so reverting would silently demote valid teknisi accounts. Keeping the
-- canonical role is the non-destructive choice; recreate a "technician" role
-- explicitly through RBAC if it is ever needed again.
SELECT 1;
