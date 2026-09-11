-- 000026_add_billing_uniqueness_indexes.up.sql
-- Idempotensi billing di level DB (F3-9) + anti-duplikat akun router (F3-10).
CREATE UNIQUE INDEX IF NOT EXISTS uq_invoices_subscription_period
    ON invoices (subscription_id, period)
    WHERE subscription_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_subscriptions_device_username
    ON subscriptions (device_id, remote_username)
    WHERE deleted_at IS NULL;
