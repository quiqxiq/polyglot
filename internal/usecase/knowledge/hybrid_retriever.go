package knowledge

import (
	"context"
	"strings"

	"github.com/quixiq/polyglot/internal/domain/knowledge"
	"github.com/quixiq/polyglot/internal/port"
)

// HybridRetriever menggabungkan hasil beberapa port.KnowledgeRetriever menjadi
// satu, dengan dedupe pada (title, content) yang identik persis.
//
// Kenapa perlu: dengan fitur embed per-dokumen (embed_to_llm), dokumen lokal
// (tidak di-embed ke AnythingLLM) hanya bisa dijawab lewat keyword retriever
// (Postgres), sementara dokumen yang di-embed dijawab lewat vector retriever
// (AnythingLLM). Kalau wiring memakai either/or seperti sebelumnya, dokumen
// lokal jadi tidak pernah ter-retrieve saat AnythingLLM aktif — hybrid
// menyatukan keduanya.
type HybridRetriever struct {
	retrievers []port.KnowledgeRetriever
}

// NewHybridRetriever membangun HybridRetriever; retriever nil diabaikan.
// Urutan array menentukan prioritas hasil (hasil retriever pertama muncul
// lebih dulu di prompt).
func NewHybridRetriever(retrievers ...port.KnowledgeRetriever) *HybridRetriever {
	nonNil := make([]port.KnowledgeRetriever, 0, len(retrievers))
	for _, r := range retrievers {
		if r != nil {
			nonNil = append(nonNil, r)
		}
	}
	return &HybridRetriever{retrievers: nonNil}
}

// Retrieve memanggil semua retriever dan menggabungkan hasilnya. Satu
// retriever error TIDAK menggagalkan retrieval (fail-open) — hasil yang
// berhasil tetap dipakai.
func (h *HybridRetriever) Retrieve(ctx context.Context, query string) ([]knowledge.KnowledgeEntry, error) {
	if strings.TrimSpace(query) == "" {
		return nil, nil
	}

	var out []knowledge.KnowledgeEntry
	seen := make(map[string]bool)
	for _, r := range h.retrievers {
		entries, err := r.Retrieve(ctx, query)
		if err != nil {
			// Fail-open: satu sumber mati jangan mematikan bot.
			continue
		}
		for i := range entries {
			e := entries[i]
			key := strings.ToLower(strings.TrimSpace(e.Title)) + "\x00" + e.Content
			if key == "\x00" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, e)
		}
	}
	return out, nil
}
