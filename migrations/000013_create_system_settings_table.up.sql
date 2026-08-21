-- SQL Migration UP: Create system_settings table and seed default configuration

CREATE TABLE IF NOT EXISTS system_settings (
    key VARCHAR(100) PRIMARY KEY,
    value TEXT NOT NULL,
    category VARCHAR(50) NOT NULL DEFAULT 'general',
    description TEXT,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed initial configuration
INSERT INTO system_settings (key, value, category, description) VALUES
    ('bot.burst_limit', '3', 'bot_rate_limit', 'Jumlah pesan maksimal dalam rentang burst sebelum dianggap spam'),
    ('bot.burst_window_secs', '5', 'bot_rate_limit', 'Rentang waktu deteksi burst spam dalam detik'),
    ('bot.mute_1h_secs', '3600', 'bot_rate_limit', 'Durasi bisukan (mute) sementara saat terdeteksi spam level 1'),
    ('bot.ban_24h_secs', '86400', 'bot_rate_limit', 'Durasi blokir (ban) saat terdeteksi spam berulang level 2'),
    ('bot.daily_chat_limit', '10', 'bot_rate_limit', 'Batas percakapan gratis harian per pelanggan'),
    ('bot.session_timeout_minutes', '30', 'bot_session', 'Batas waktu kedaluwarsa sesi obrolan tanpa interaksi (menit)'),
    ('bot.sliding_window_size', '10', 'bot_session', 'Jumlah pesan riwayat terakhir yang dikirim ke LLM'),
    ('bot.llm_max_output_tokens', '1024', 'bot_session', 'Batas maksimal output token respon LLM'),
    ('bot.whitelist_all_staff', 'true', 'bot_whitelist', 'Otomatis bebaskan limit untuk seluruh nomor HP di tabel users'),
    ('bot.custom_whitelist_phones', '', 'bot_whitelist', 'Nomor WhatsApp tambahan yang dibebaskan dari limit (pisahkan koma)')
ON CONFLICT (key) DO NOTHING;
