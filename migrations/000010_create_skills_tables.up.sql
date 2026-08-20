CREATE TABLE IF NOT EXISTS bot_skills (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT DEFAULT '',
    is_enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bot_skills_slug ON bot_skills(slug);
CREATE INDEX IF NOT EXISTS idx_bot_skills_enabled ON bot_skills(is_enabled);

CREATE TABLE IF NOT EXISTS bot_skill_files (
    id SERIAL PRIMARY KEY,
    skill_id INT NOT NULL REFERENCES bot_skills(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    content TEXT NOT NULL DEFAULT '',
    is_reference BOOLEAN DEFAULT FALSE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bot_skill_files_skill_id ON bot_skill_files(skill_id);
CREATE INDEX IF NOT EXISTS idx_bot_skill_files_path ON bot_skill_files(file_path);

CREATE TABLE IF NOT EXISTS bot_global_prompts (
    id SERIAL PRIMARY KEY,
    key VARCHAR(50) UNIQUE NOT NULL DEFAULT 'default',
    content TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
