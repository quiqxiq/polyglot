package port

import "context"

// KnowledgeDocumentManager defines the write-side contract for syncing a
// knowledge document to the AnythingLLM vector store. Ini sisi tulis dari
// fitur admin knowledge: dipakai usecase knowledge (manage_document.go) untuk
// meng-embed, re-embed, dan menghapus dokumen per knowledge entry. Sisi baca
// (retrieval untuk bot) tetap lewat KnowledgeRetriever.
type KnowledgeDocumentManager interface {
	// UpsertDocument meng-upload isi markdown satu dokumen ke AnythingLLM dan
	// mengembalikan doc name JSON terbaru (path relatif storage, mis.
	// "custom-documents/raw-...json") untuk disimpan di kolom
	// knowledge_entries.anythingllm_doc_name.
	//
	// docName adalah doc name JSON yang terakhir berhasil di-embed ("" kalau
	// belum pernah). Karena AnythingLLM tidak punya API update content — setiap
	// raw-text upload menghasilkan dokumen baru dengan nama UUID acak — doc
	// lama dihapus dulu sebelum upload baru (semantik replace).
	UpsertDocument(ctx context.Context, docName, title, markdown string) (string, error)

	// DeleteDocument menghapus satu dokumen dari AnythingLLM. docName harus
	// path relatif storage yang dikembalikan UpsertDocument; kosong = no-op
	// (dokumen belum pernah di-embed).
	DeleteDocument(ctx context.Context, docName string) error
}
