-- 000040_payment_ops.down.sql
DELETE FROM notification_templates WHERE id = 'nt-payment-rejected';

DELETE FROM system_settings WHERE key IN (
    'isp.gateway_fee_expense_category_id','isp.mass_payment_enabled'
);

DROP TABLE IF EXISTS bulk_payment_items;
DROP TABLE IF EXISTS bulk_payment_batches;
DROP TABLE IF EXISTS gateway_settlements;

ALTER TABLE payment_methods
    DROP COLUMN IF EXISTS bank_name,
    DROP COLUMN IF EXISTS account_number,
    DROP COLUMN IF EXISTS account_holder,
    DROP COLUMN IF EXISTS qris_image_url,
    DROP COLUMN IF EXISTS instructions,
    DROP COLUMN IF EXISTS sort_order,
    DROP COLUMN IF EXISTS updated_at;

DELETE FROM cash_categories WHERE id = 'cc-gateway-fee';
