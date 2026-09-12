-- 000031_status_check_constraints.down.sql
ALTER TABLE payments DROP CONSTRAINT IF EXISTS chk_payment_scan_method;
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS chk_invoice_status;
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS chk_sub_service_type;
ALTER TABLE subscriptions DROP CONSTRAINT IF EXISTS chk_sub_status;
