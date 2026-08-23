-- 000018_isp_settings_and_provision_status.up.sql
-- Fase 3.5: konfigurasi billing/isolir dinamis via system_settings
-- (bukan konstanta kode) + kolom status provisi router pada subscriptions.

ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS provision_status VARCHAR(20) NOT NULL DEFAULT 'NONE',
    ADD COLUMN IF NOT EXISTS router_profile VARCHAR(100);
-- provision_status: NONE | PENDING | OK | FAILED — jejak sinkronisasi akun
-- ke router. router_profile: nama profil aktif akhir (basis restore).

INSERT INTO system_settings (key, value, category, description) VALUES
    ('isp.billing_cycle_start_day',  '1',     'isp_billing',
     'Tanggal mulai periode tagihan dalam sebulan (1-28)'),
    ('isp.billing_due_days',         '20',    'isp_billing',
     'Jatuh tempo tagihan: hari-X kalender setelah periode terbit'),
    ('isp.auto_isolate',             'true',  'isp_isolation',
     'Aktifkan isolir otomatis saat tagihan melewati jatuh tempo'),
    ('isp.isolate_grace_days',       '3',     'isp_isolation',
     'Hari toleransi setelah jatuh tempo sebelum diisolir'),
    ('isp.pppoe_isolir_profile',     'isolir','isp_isolation',
     'Nama profil PPPoE isolir di router (akun dipindah ke profil ini)'),
    ('isp.hotspot_isolir_profile',   'isolir','isp_isolation',
     'Nama profil hotspot isolir di router'),
    ('isp.payment_redirect_url',     '',      'isp_isolation',
     'URL halaman pembayaran tujuan redirect dst-nat saat isolir (host:port)'),
    ('isp.suspend_after_days',       '90',    'isp_isolation',
     'Tunggakan hari sebelum ISOLATED otomatis jadi SUSPENDED (0 = off, admin manual)'),
    ('isp.isolir_address_list',      'ISOLIR_USERS', 'isp_isolation',
     'Nama address-list penanda pelanggan terisolir untuk rule redirect')
ON CONFLICT (key) DO NOTHING;
