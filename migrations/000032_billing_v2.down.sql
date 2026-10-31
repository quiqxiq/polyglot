-- 000032_billing_v2.down.sql
-- Reverse 000032_billing_v2.up.sql.

-- Seed settings.
DELETE FROM system_settings WHERE key IN (
    'isp.first_invoice_due_days',
    'isp.bill_isolated',
    'isp.bill_suspended',
    'isp.gateway_fee_bearer',
    'isp.gateway_markup_amount'
);

-- Tabel baru (urutan: turunan dulu, independen tidak masalah).
DROP TABLE IF EXISTS registration_events;
DROP TABLE IF EXISTS payment_submissions;

-- registrations: kolom PSB v2 + kembalikan constraint status lama.
ALTER TABLE registrations
    DROP COLUMN IF EXISTS sales_agent_id,
    DROP COLUMN IF EXISTS discount_amount,
    DROP COLUMN IF EXISTS ont_serial_number,
    DROP COLUMN IF EXISTS ont_mac_address,
    DROP COLUMN IF EXISTS installation_photo_url,
    DROP COLUMN IF EXISTS installation_started_at,
    DROP COLUMN IF EXISTS public_token,
    DROP COLUMN IF EXISTS estimated_cable_length_meter,
    DROP COLUMN IF EXISTS survey_notes,
    DROP COLUMN IF EXISTS survey_status,
    DROP COLUMN IF EXISTS surveyor_id,
    DROP COLUMN IF EXISTS survey_time,
    DROP COLUMN IF EXISTS survey_date,
    DROP COLUMN IF EXISTS house_photo_url,
    DROP COLUMN IF EXISTS id_card_photo_url,
    DROP COLUMN IF EXISTS nik;

ALTER TABLE registrations
    DROP CONSTRAINT IF EXISTS chk_reg_status,
    ADD CONSTRAINT chk_reg_status CHECK (status IN ('PENDING','APPROVED','INSTALLED','ACTIVE','REJECTED','CANCELLED'));

-- invoices: constraint unik idempotensi + kolom baru.
ALTER TABLE invoices
    DROP CONSTRAINT IF EXISTS uq_invoices_sub_period_type,
    DROP COLUMN IF EXISTS issued_at,
    DROP COLUMN IF EXISTS credit_note_of,
    DROP COLUMN IF EXISTS promo_code,
    DROP COLUMN IF EXISTS discount_amount,
    DROP COLUMN IF EXISTS discount_value,
    DROP COLUMN IF EXISTS discount_type,
    DROP COLUMN IF EXISTS period_days,
    DROP COLUMN IF EXISTS prorated_days,
    DROP COLUMN IF EXISTS period_end,
    DROP COLUMN IF EXISTS period_start,
    DROP COLUMN IF EXISTS invoice_type;

-- subscriptions.
ALTER TABLE subscriptions
    DROP COLUMN IF EXISTS provision_attempts,
    DROP COLUMN IF EXISTS last_provision_error,
    DROP COLUMN IF EXISTS next_due_date,
    DROP COLUMN IF EXISTS paid_through,
    DROP COLUMN IF EXISTS billing_anchor_day;

-- service_plans.
DROP INDEX IF EXISTS idx_service_plans_profile;
ALTER TABLE service_plans
    DROP COLUMN IF EXISTS version,
    DROP COLUMN IF EXISTS tax_mode,
    DROP COLUMN IF EXISTS prorate_mode,
    DROP COLUMN IF EXISTS router_profile_name;
