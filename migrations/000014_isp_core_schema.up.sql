-- 000014_isp_core_schema.up.sql
--
-- ISP core schema (Fase ISP-1): service_plans + registrations canonical,
-- customers enriched, subscriptions rebuilt as MAPPING-ONLY table.
--
-- Prinsip mapping-only (docs/database-schema.md §7.1):
--   Router Mikrotik = source of truth detail akun (password, IP terpasang).
--   DB hanya menyimpan identitas bisnis + mapping:
--     device_id       → router tempat akun diprovisi
--     remote_username → kunci natural akun di device (PPPoE secret / hotspot user)
--     remote_id       → RouterOS .id untuk operasi set/remove tanpa print
--   Password akun TIDAK disimpan di sini — via SecretVault (AES-GCM),
--   pola credentials tabel 000001.
--
-- Skeleton lama (migrasi 000006) di-drop karena belum berisi data produksi.

DROP TABLE IF EXISTS subscriptions CASCADE;
DROP TABLE IF EXISTS plans CASCADE;

-- ── customers: diperkaya (struktur 000006 dipertahankan) ─────────────────
ALTER TABLE customers
    ADD COLUMN IF NOT EXISTS customer_code VARCHAR(30),
    ADD COLUMN IF NOT EXISTS latitude      DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS longitude     DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS notes         TEXT,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS uq_customers_code
    ON customers (tenant_id, customer_code)
    WHERE deleted_at IS NULL AND customer_code <> '';
CREATE INDEX IF NOT EXISTS idx_customers_phone
    ON customers (phone) WHERE deleted_at IS NULL;

-- ── service_plans: master paket, type-driven fields ──────────────────────
-- Kolom router-specific adalah REFERENSI NAMA objek di router (ip pool,
-- parent queue, address list) — bukan snapshot state. Rate-limit final
-- dihitung dari rate_down/up_kbps saat provisioning profile.
CREATE TABLE service_plans (
    id             TEXT PRIMARY KEY,
    tenant_id      TEXT NOT NULL DEFAULT 'tenant-default',
    name           VARCHAR(100) NOT NULL,
    service_type   VARCHAR(20)  NOT NULL CHECK (service_type IN ('PPPOE', 'HOTSPOT')),
    rate_down_kbps INT NOT NULL CHECK (rate_down_kbps > 0),
    rate_up_kbps   INT NOT NULL CHECK (rate_up_kbps > 0),
    price          NUMERIC(15,2) NOT NULL CHECK (price >= 0),
    ip_pool_name   VARCHAR(50),          -- PPPOE: nama /ip pool untuk remote-address
    parent_queue   VARCHAR(50),          -- PPPOE opsional: parent queue sederhana
    address_list   VARCHAR(50),          -- PPPOE opsional: address-list firewall
    shared_users   INT NOT NULL DEFAULT 1 CHECK (shared_users > 0), -- HOTSPOT
    is_active      BOOLEAN NOT NULL DEFAULT TRUE,
    description    TEXT,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_service_plans_name UNIQUE (tenant_id, name)
);

CREATE INDEX idx_service_plans_type   ON service_plans (service_type);
CREATE INDEX idx_service_plans_active ON service_plans (is_active);

-- ── subscriptions: mapping-only ke router Mikrotik ───────────────────────
CREATE TABLE subscriptions (
    id               TEXT PRIMARY KEY,
    tenant_id        TEXT NOT NULL DEFAULT 'tenant-default',
    customer_id      TEXT NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    plan_id          TEXT NOT NULL REFERENCES service_plans(id) ON DELETE RESTRICT,
    -- MAPPING provisioning
    device_id        UUID REFERENCES devices(id) ON DELETE SET NULL,
    service_type     VARCHAR(20) NOT NULL CHECK (service_type IN ('PPPOE', 'HOTSPOT')),
    remote_username  TEXT,
    remote_id        TEXT,
    -- penagihan dasar (billing automation menyusul fase berikutnya)
    billing_day      INT NOT NULL DEFAULT 1 CHECK (billing_day BETWEEN 1 AND 28),
    -- lifecycle
    status           VARCHAR(20) NOT NULL DEFAULT 'PENDING_PROVISION'
                     CHECK (status IN ('PENDING_PROVISION','ACTIVE','ISOLATED','SUSPENDED','TERMINATED')),
    start_date       DATE NOT NULL DEFAULT CURRENT_DATE,
    isolated_at      TIMESTAMPTZ,
    isolation_reason TEXT,
    notes            TEXT,
    deleted_at       TIMESTAMPTZ,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_subscriptions_customer ON subscriptions (customer_id);
CREATE INDEX idx_subscriptions_device   ON subscriptions (device_id);
CREATE INDEX idx_subscriptions_status   ON subscriptions (status) WHERE deleted_at IS NULL;
-- Satu akun device hanya boleh dipakai satu langganan aktif.
CREATE UNIQUE INDEX uq_subscriptions_remote
    ON subscriptions (device_id, remote_username)
    WHERE remote_username IS NOT NULL AND deleted_at IS NULL;

CREATE TRIGGER trg_service_plans_upd BEFORE UPDATE ON service_plans
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE TRIGGER trg_subscriptions_upd BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── registrations: alur pendaftaran pelanggan baru ───────────────────────
-- PENDING → APPROVED → INSTALLED (provisioning dipicu di sini) → ACTIVE
-- atau REJECTED / CANCELLED.
CREATE TABLE registrations (
    id                     TEXT PRIMARY KEY,
    tenant_id              TEXT NOT NULL DEFAULT 'tenant-default',
    registration_no        VARCHAR(30) UNIQUE NOT NULL,
    plan_id                TEXT NOT NULL REFERENCES service_plans(id) ON DELETE RESTRICT,
    -- data pemohon ringkas (tanpa NIK/KTP)
    full_name              VARCHAR(100) NOT NULL,
    phone                  VARCHAR(20)  NOT NULL,
    address                TEXT         NOT NULL,
    latitude               DOUBLE PRECISION,
    longitude              DOUBLE PRECISION,
    notes                  TEXT,
    -- alur status
    status                 VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    reviewed_by            BIGINT REFERENCES users(id) ON DELETE SET NULL,
    reviewed_at            TIMESTAMPTZ,
    admin_notes            TEXT,
    scheduled_install_date DATE,
    assigned_technician_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    -- hasil pemasangan teknisi lapangan; device_id dipilih teknisi
    device_id              UUID REFERENCES devices(id) ON DELETE SET NULL,
    installed_at           TIMESTAMPTZ,
    technician_notes       TEXT,
    -- hasil konversi menjadi pelanggan aktif
    customer_id            TEXT REFERENCES customers(id) ON DELETE SET NULL,
    subscription_id        TEXT REFERENCES subscriptions(id) ON DELETE SET NULL,
    rejected_at            TIMESTAMPTZ,
    rejected_reason        TEXT,
    cancelled_at           TIMESTAMPTZ,
    cancel_reason          TEXT,
    created_at             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT chk_reg_status CHECK (status IN
        ('PENDING','APPROVED','INSTALLED','ACTIVE','REJECTED','CANCELLED'))
);

CREATE INDEX idx_registrations_status ON registrations (status, created_at DESC);
CREATE INDEX idx_registrations_phone  ON registrations (phone);
CREATE INDEX idx_registrations_tech   ON registrations (assigned_technician_id, status);
-- Cegah pendaftaran ganda dengan nomor yang sama selama masih dalam proses.
CREATE UNIQUE INDEX uq_registrations_active_phone ON registrations (phone)
    WHERE status IN ('PENDING', 'APPROVED');

CREATE TRIGGER trg_registrations_upd BEFORE UPDATE ON registrations
    FOR EACH ROW EXECUTE FUNCTION set_updated_at();

-- ── secrets: vault rahasia umum (password akun PPPoE/hotspot permanent) ──
-- Pola AES-GCM sama dengan tabel credentials (000001); key = kunci logis
-- bebas (konvensi: "subscription:<id>:password").
CREATE TABLE secrets (
    key        TEXT PRIMARY KEY,
    ciphertext BYTEA NOT NULL,
    nonce      BYTEA NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ── seed config isolir (system_settings sudah ada sejak 000013) ──────────
INSERT INTO system_settings (key, value, category, description) VALUES
    ('isolir.profile_name',   'isolir',                          'isolir', 'Nama profile suspensi PPPoE di router'),
    ('isolir.pool_name',      'pool-isolir',                     'isolir', 'Nama /ip pool khusus pelanggan terisolir'),
    ('isolir.pool_range',     '172.16.99.10-172.16.99.254',      'isolir', 'Range IP pool isolir'),
    ('isolir.portal_ip',      '',                                'isolir', 'IP portal pembayaran tujuan redirect (diisi saat portal siap)'),
    ('isolir.portal_http_port','8080',                           'isolir', 'Port listen portal pembayaran'),
    ('isolir.redirect_ports', '80,443',                          'isolir', 'Port trafik pelanggan yang di-redirect ke portal')
ON CONFLICT (key) DO NOTHING;
