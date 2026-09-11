-- 000027_seed_refund_category.up.sql
-- Kategori kas untuk koreksi/refund pembatalan invoice (F3-6).
INSERT INTO cash_categories (id, tenant_id, name, type, is_active) VALUES
    ('cc-refund', 'tenant-default', 'Refund/Koreksi', 'EXPENSE', TRUE)
ON CONFLICT (id) DO NOTHING;
