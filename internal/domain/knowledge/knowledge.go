package knowledge

import "time"

// KnowledgeEntry represents a single FAQ/procedure entry in the knowledge base.
type KnowledgeEntry struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      string    `json:"tags"` // comma-separated tags for keyword matching (v1)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
