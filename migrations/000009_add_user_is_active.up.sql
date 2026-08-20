-- SQL Migration UP: User is_active (akun aktif/nonaktif)

-- `is_active` : true = akun bisa login, false = akun di-disable (Login
--                ditolak). Kolom baru — user management (UserService) butuh
--                cara menonaktifkan akun tanpa menghapus datanya.
ALTER TABLE users
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;
