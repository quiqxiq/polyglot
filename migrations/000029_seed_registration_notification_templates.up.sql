-- 000029_seed_registration_notification_templates.up.sql
-- Template notifikasi alur pemasangan/penolakan/pembatalan registrasi (F6-5).
INSERT INTO notification_templates (id, template_key, name, content, variables_json) VALUES
    ('nt-install-completed', 'INSTALLATION_COMPLETED',
     'Pemasangan Selesai',
     'Halo {{full_name}}, pemasangan layanan di {{address}} telah selesai. Akun Anda akan segera aktif.',
     '["full_name","address"]'),
    ('nt-reg-rejected', 'REGISTRATION_REJECTED',
     'Pendaftaran Ditolak',
     'Halo {{full_name}}, pendaftaran Anda tidak dapat diproses. Alasan: {{reason}}. Hubungi kami untuk informasi lebih lanjut.',
     '["full_name","reason"]'),
    ('nt-reg-cancelled', 'REGISTRATION_CANCELLED',
     'Pendaftaran Dibatalkan',
     'Halo {{full_name}}, pendaftaran Anda telah dibatalkan. Alasan: {{reason}}.',
     '["full_name","reason"]')
ON CONFLICT (id) DO NOTHING;
