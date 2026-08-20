package skill

import (
	"time"
)

// Skill merepresentasikan satu paket kemampuan bot / agent modular (SKILL.md).
type Skill struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`          // Nama unik / identifier folder (e.g. "troubleshoot-los-onu")
	Description   string            `json:"description"`   // Penjelasan ringkas pemicu skill
	Content       string            `json:"content"`       // Markdown body (tanpa frontmatter)
	License       string            `json:"license,omitempty"`
	Compatibility string            `json:"compatibility,omitempty"`
	AllowedTools  string            `json:"allowed_tools,omitempty"` // Daftar tool yang diizinkan (e.g. "bash, mcp_device")
	Metadata      map[string]string `json:"metadata,omitempty"`      // Key-value metadata tambahan
	ReadOnly      bool              `json:"read_only"`               // True jika bersumber dari repositori Git
	SourceType    string            `json:"source_type"`             // "inline" atau "git"
	SourceURL     string            `json:"source_url,omitempty"`    // URL git jika bersumber dari git
	Version       string            `json:"version,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// SkillResource merepresentasikan file aset atau skrip pendukung di dalam folder skill (misal: scripts/helper.rsc).
type SkillResource struct {
	Path     string    `json:"path"`      // Relative path di dalam folder skill (e.g. "scripts/diag.rsc")
	Name     string    `json:"name"`      // Nama file dasar (e.g. "diag.rsc")
	Type     string    `json:"type"`      // "script", "reference", atau "asset"
	Size     int64     `json:"size"`      // Ukuran file dalam bytes
	MimeType string    `json:"mime_type"` // MIME type file
	Readable bool      `json:"readable"`  // True jika file berupa teks yang dapat dibaca manusia
	Modified time.Time `json:"modified"`
}

// ResourceContent merepresentasikan isi data dari file resource tertentu.
type ResourceContent struct {
	Content  string `json:"content"`   // Teks biasa atau string base64 jika biner
	Encoding string `json:"encoding"`  // "raw" atau "base64"
	MimeType string `json:"mime_type"`
	Size     int64  `json:"size"`
}

// GitRepoInfo mendeskripsikan repositori Git yang dikonfigurasikan sebagai sumber skill.
type GitRepoInfo struct {
	ID           string    `json:"id"`
	URL          string    `json:"url"`
	Name         string    `json:"name"`
	Enabled      bool      `json:"enabled"`
	LastSyncedAt time.Time `json:"last_synced_at,omitempty"`
}

// SkillInfo adalah representasi ringan skill untuk transfer data ke runtime agent / LLM prompt.
type SkillInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content,omitempty"` // Konten lengkap untuk mode injeksi prompt
}

// SkillMetadataRecord merepresentasikan catatan metadata skill yang disimpan di PostgreSQL.
type SkillMetadataRecord struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id,omitempty"`
	Name       string    `json:"name"`
	Definition string    `json:"definition,omitempty"` // Ringkasan deskripsi (dipotong max 500 karakter)
	SourceType string    `json:"source_type"`          // "inline", "git"
	SourceURL  string    `json:"source_url,omitempty"`
	Version    string    `json:"version,omitempty"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
