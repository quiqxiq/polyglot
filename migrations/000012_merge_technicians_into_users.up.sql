-- SQL Migration UP: Merge Technicians into Users table

ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name VARCHAR(255);
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone_number VARCHAR(50);
ALTER TABLE users ADD COLUMN IF NOT EXISTS specialization VARCHAR(255);

-- Migrasi data dari tabel technicians ke tabel users jika tabel technicians ada
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'technicians') THEN
        INSERT INTO users (username, email, password_hash, role, full_name, phone_number, specialization, is_active, tenant_id, created_at, updated_at)
        SELECT 
            t.username,
            COALESCE(t.username || '@gnet.local', 'tech_' || t.id || '@gnet.local'),
            '$2a$10$defaultHashPlaceholderForTechnicianAccountOnly1234567890',
            'teknisi',
            t.full_name,
            t.phone_number,
            t.specialization,
            t.is_active,
            'tenant-default',
            t.created_at,
            t.updated_at
        FROM technicians t
        ON CONFLICT (username) DO UPDATE SET
            full_name = EXCLUDED.full_name,
            phone_number = EXCLUDED.phone_number,
            specialization = EXCLUDED.specialization;

        DROP TABLE IF EXISTS technicians;
    END IF;
END $$;
