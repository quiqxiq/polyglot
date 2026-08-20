-- Add optional ssh_port column to devices table (default 22 for SSH PTY connections)
ALTER TABLE devices ADD COLUMN IF NOT EXISTS ssh_port INT DEFAULT 22;
