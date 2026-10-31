-- 000039_debts.down.sql
DELETE FROM notification_templates WHERE id IN ('nt-debt-reminder','nt-debt-paid');

DELETE FROM system_settings WHERE key IN (
    'isp.debt_income_category_id','isp.debt_reminder_enabled','isp.debt_reminder_days'
);

DROP TABLE IF EXISTS debt_payments;
DROP TABLE IF EXISTS customer_debts;

DELETE FROM cash_categories WHERE id = 'cc-debt-income';
