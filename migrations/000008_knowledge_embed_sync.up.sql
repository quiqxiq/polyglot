-- SQL Migration UP: Knowledge Embed Sync (per-document toggle to AnythingLLM)

-- `category` ditambahkan karena proto KnowledgeItem sudah punya field category
-- sejak awal tapi tabel 000002 belum menyediakan kolomnya.
-- `embed_to_llm`    : flag per-dokumen — true berarti dokumen ikut di-embed
--                      ke vector store AnythingLLM, false = murni lokal (admin).
-- `embed_status`    : none | pending | embedded | failed (lihat
--                      internal/domain/knowledge/knowledge.go).
-- `anythingllm_doc_name` : nama dokumen JSON di AnythingLLM yang terakhir
--                      berhasil di-embed, dipakai untuk delete/re-embed.
ALTER TABLE knowledge_entries
    ADD COLUMN category VARCHAR(100) NOT NULL DEFAULT 'umum',
    ADD COLUMN embed_to_llm BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN embed_status VARCHAR(20) NOT NULL DEFAULT 'none',
    ADD COLUMN anythingllm_doc_name VARCHAR(255);

CREATE INDEX idx_knowledge_embed_status ON knowledge_entries(embed_status);
