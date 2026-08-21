-- SQL Migration DOWN: Rollback merged technicians columns from users

ALTER TABLE users DROP COLUMN IF EXISTS full_name;
ALTER TABLE users DROP COLUMN IF EXISTS phone_number;
ALTER TABLE users DROP COLUMN IF EXISTS specialization;

CREATE TABLE IF NOT EXISTS technicians (
    id SERIAL PRIMARY KEY,
    full_name VARCHAR(255) NOT NULL,
    username VARCHAR(100) NOT NULL UNIQUE,
    phone_number VARCHAR(50) NOT NULL,
    specialization VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
