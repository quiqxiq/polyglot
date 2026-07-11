-- Lepas ketiga FK penutup siklus dulu (urutan terbalik dari .up),
-- sebelum drop tabel yang dibuat migration ini.
ALTER TABLE command_audit_log DROP CONSTRAINT fk_command_audit_hitl;
ALTER TABLE hitl_approvals DROP CONSTRAINT fk_hitl_mcp_invocation;
ALTER TABLE mcp_tool_invocations DROP CONSTRAINT fk_mcp_invocation_conversation;

DROP TABLE ai_messages;
DROP TABLE ai_conversations;
