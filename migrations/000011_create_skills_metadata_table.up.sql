CREATE TABLE IF NOT EXISTS skills_metadata (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) DEFAULT '',
    name VARCHAR(255) NOT NULL,
    definition TEXT DEFAULT '',
    source_type VARCHAR(32) NOT NULL DEFAULT 'inline',
    source_url VARCHAR(512) DEFAULT '',
    version VARCHAR(64) DEFAULT '',
    enabled BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_skills_metadata_user_id ON skills_metadata(user_id);
CREATE INDEX IF NOT EXISTS idx_skills_metadata_name ON skills_metadata(name);
CREATE INDEX IF NOT EXISTS idx_skills_metadata_enabled ON skills_metadata(enabled);
