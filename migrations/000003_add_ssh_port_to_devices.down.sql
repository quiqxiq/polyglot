-- Drop optional ssh_port column from devices table
ALTER TABLE devices DROP COLUMN IF EXISTS ssh_port;
