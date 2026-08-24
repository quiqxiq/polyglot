-- 000020_add_remote_address_pool.down.sql
ALTER TABLE service_plans
    DROP COLUMN IF EXISTS remote_address_pool;
