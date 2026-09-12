-- 000031_status_check_constraints.up.sql
-- CHECK constraint status/enum agar nilai liar ditolak di level DB (F6-12).
ALTER TABLE subscriptions
    ADD CONSTRAINT chk_sub_status CHECK (status IN ('ACTIVE','ISOLATED','SUSPENDED','TERMINATED','PENDING','CANCELLED','EXPIRED'));
ALTER TABLE subscriptions
    ADD CONSTRAINT chk_sub_service_type CHECK (service_type IN ('PPPOE','HOTSPOT','DEDICATED'));
ALTER TABLE invoices
    ADD CONSTRAINT chk_invoice_status CHECK (status IN ('UNPAID','PARTIAL','PAID','OVERDUE','CANCELLED'));
ALTER TABLE payments
    ADD CONSTRAINT chk_payment_scan_method CHECK (scan_method IN ('QR_SCAN','CODE_INPUT','MANUAL','PAYMENT_GATEWAY'));
