-- SQL Migration DOWN: User is_active

ALTER TABLE users
    DROP COLUMN IF EXISTS is_active;
