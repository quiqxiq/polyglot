-- 000033_payment_reversal.down.sql
DELETE FROM system_settings WHERE key IN (
    'isp.manual_payment_cash_account_id',
    'isp.manual_payment_income_category_id',
    'isp.manual_payment_method_id',
    'isp.gateway_fee_percent'
);

DROP INDEX IF EXISTS idx_payments_voided;

ALTER TABLE payments
    DROP COLUMN IF EXISTS voided_by,
    DROP COLUMN IF EXISTS void_reason,
    DROP COLUMN IF EXISTS voided_at;
