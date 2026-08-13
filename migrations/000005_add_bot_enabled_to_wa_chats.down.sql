-- SQL Migration DOWN: Per-chat bot control
ALTER TABLE wa_chats DROP COLUMN IF EXISTS bot_enabled;
