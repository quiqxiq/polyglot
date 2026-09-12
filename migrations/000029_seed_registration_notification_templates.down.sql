-- 000029_seed_registration_notification_templates.down.sql
DELETE FROM notification_templates WHERE id IN (
    'nt-install-completed', 'nt-reg-rejected', 'nt-reg-cancelled'
);
