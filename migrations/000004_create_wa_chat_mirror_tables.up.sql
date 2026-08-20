-- SQL Migration UP: Create WhatsApp Chat Mirror Tables (Inbox)
--
-- wa_chats / wa_messages adalah mirror dari percakapan WhatsApp nyata per
-- perangkat (session). Berbeda dari `conversations`/`messages` (yang hanya
-- mencatat chat yang ditangani bot), tabel ini mencatat SEMUA chat dan pesan
-- yang masuk/keluar agar halaman Inbox menampilkan seluruh percakapan.

CREATE TABLE IF NOT EXISTS wa_chats (
    id SERIAL PRIMARY KEY,
    session_id INT NOT NULL REFERENCES wa_sessions(id) ON DELETE CASCADE,
    chat_jid VARCHAR(255) NOT NULL,
    display_name VARCHAR(255),
    is_group BOOLEAN NOT NULL DEFAULT FALSE,
    last_message_id VARCHAR(255),
    last_message_preview TEXT,
    last_message_time TIMESTAMP WITH TIME ZONE,
    unread_count INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_wa_chats_session_jid UNIQUE (session_id, chat_jid)
);

CREATE INDEX IF NOT EXISTS idx_wa_chats_session_last_time ON wa_chats(session_id, last_message_time DESC);

CREATE TABLE IF NOT EXISTS wa_messages (
    id SERIAL PRIMARY KEY,
    session_id INT NOT NULL REFERENCES wa_sessions(id) ON DELETE CASCADE,
    chat_jid VARCHAR(255) NOT NULL,
    wa_message_id VARCHAR(255) NOT NULL,
    sender_jid VARCHAR(255),
    sender_name VARCHAR(255),
    content TEXT,
    media_type VARCHAR(50) NOT NULL DEFAULT 'text',
    is_from_me BOOLEAN NOT NULL DEFAULT FALSE,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    timestamp TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT uq_wa_messages_session_wa_id UNIQUE (session_id, wa_message_id)
);

CREATE INDEX IF NOT EXISTS idx_wa_messages_session_chat_time ON wa_messages(session_id, chat_jid, timestamp DESC);
