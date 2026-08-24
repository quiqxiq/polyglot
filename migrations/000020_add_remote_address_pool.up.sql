-- 000020_add_remote_address_pool.up.sql
-- Remote address pool untuk plan PPPoE/Dedicated: pelanggan dial mendapat IP
-- dari pool ini (kolom remote-address pada /ppp/profile RouterOS).
ALTER TABLE service_plans
    ADD COLUMN IF NOT EXISTS remote_address_pool VARCHAR(50);
