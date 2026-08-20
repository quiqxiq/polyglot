-- SQL Migration UP: Per-chat bot control
--
-- bot_enabled menandai apakah bot auto-reply aktif untuk SATU chat tertentu
-- (per session + chat_jid). Default TRUE (mengikuti switch bot perangkat).
-- Berguna untuk mode "ambil alih": agen bisa menonaktifkan bot hanya di satu
-- percakapan tanpa mengganggu perangkat lain.

ALTER TABLE wa_chats ADD COLUMN IF NOT EXISTS bot_enabled BOOLEAN NOT NULL DEFAULT TRUE;
