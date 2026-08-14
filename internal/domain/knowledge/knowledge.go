package knowledge

import "time"

// Status embed per-dokumen ke AnythingLLM. Disimpan di kolom
// `knowledge_entries.embed_status` (migrasi 000008) dan diekspos ke frontend
// lewat `KnowledgeItem.embed_status` untuk badge status.
const (
	// EmbedStatusNone berarti dokumen tidak (dan tidak akan) di-embed ke
	// AnythingLLM — murni knowledge lokal yang hanya dipakai keyword retriever.
	EmbedStatusNone = "none"
	// EmbedStatusPending berarti embed diminta (embed_to_llm = true) tapi
	// sinkronisasi ke AnythingLLM belum selesai.
	EmbedStatusPending = "pending"
	// EmbedStatusEmbedded berarti dokumen aktif di vector store AnythingLLM
	// dan ikut dipakai bot lewat vector retriever.
	EmbedStatusEmbedded = "embedded"
	// EmbedStatusFailed berarti sinkronisasi gagal — dokumen tetap aman di
	// Postgres, status bisa di-retry dari UI admin.
	EmbedStatusFailed = "failed"
)

// KnowledgeEntry represents a single FAQ/procedure entry in the knowledge base.
type KnowledgeEntry struct {
	ID       uint   `json:"id"`
	Title    string `json:"title"`
	Content  string `json:"content"`
	Category string `json:"category"`
	Tags     string `json:"tags"` // comma-separated tags for keyword matching (v1)
	// EmbedToLLM menandai apakah dokumen ini ikut di-sync ke AnythingLLM
	// (vector store). False = dokumen hanya tampil di admin dashboard dan
	// dipakai keyword retriever lokal.
	EmbedToLLM bool `json:"embed_to_llm"`
	// EmbedStatus adalah status sinkronisasi terakhir (lihat konstanta di atas).
	EmbedStatus string `json:"embed_status"`
	// AnythingLLMDocName adalah nama dokumen JSON di AnythingLLM yang terakhir
	// berhasil di-embed — dipakai untuk delete/re-embed saat dokumen dihapus
	// atau isinya diubah.
	AnythingLLMDocName string    `json:"anythingllm_doc_name"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
