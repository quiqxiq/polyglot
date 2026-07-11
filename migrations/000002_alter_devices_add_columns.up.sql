-- Kolom tambahan pada devices (000001) untuk kebutuhan business layer:
-- lokasi fisik, referensi vault kredensial, dan status aktif/nonaktif.
ALTER TABLE devices ADD COLUMN site_name TEXT;
ALTER TABLE devices ADD COLUMN credential_vault_ref TEXT;
ALTER TABLE devices ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT true;
