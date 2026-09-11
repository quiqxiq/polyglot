-- 000026_add_billing_uniqueness_indexes.down.sql
DROP INDEX IF EXISTS uq_subscriptions_device_username;
DROP INDEX IF EXISTS uq_invoices_subscription_period;
