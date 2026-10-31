-- 000038_billing_ops_v3.down.sql
DELETE FROM notification_templates WHERE id IN ('nt-invoice-created','nt-bill-overdue');

DELETE FROM system_settings WHERE key IN (
    'isp.billing_mode','isp.closing_date','isp.unpaid_days_display',
    'isp.prepay_invoice_lead_days','isp.tax_percent',
    'isp.reminder_pre_offsets','isp.reminder_post_offsets',
    'isp.gateway_sync_stale_minutes'
);

DROP SEQUENCE IF EXISTS invoice_no_seq;

ALTER TABLE invoices DROP COLUMN IF EXISTS tax_percent;
