ALTER TABLE provisioning_sync_log DROP CONSTRAINT fk_sync_log_command_audit;

DROP TABLE command_audit_log;
