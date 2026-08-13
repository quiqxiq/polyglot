-- 000007_add_message_status: menambah kolom status pengiriman pesan keluar
-- pada wa_messages. Nilai mengikuti konvensi WhatsApp: "sent" (✓ terkirim),
-- "delivered" (✓✓ sampai di device penerima), "read" (✓✓ biru dibaca).
-- Pesan masuk (is_from_me=false) tidak memakai kolom ini (default sent).
ALTER TABLE wa_messages
    ADD COLUMN IF NOT EXISTS status VARCHAR(20) NOT NULL DEFAULT 'sent';
