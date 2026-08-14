-- SQL Migration DOWN: Knowledge Embed Sync

DROP INDEX IF EXISTS idx_knowledge_embed_status;

ALTER TABLE knowledge_entries
    DROP COLUMN IF EXISTS category,
    DROP COLUMN IF EXISTS embed_to_llm,
    DROP COLUMN IF EXISTS embed_status,
    DROP COLUMN IF EXISTS anythingllm_doc_name;
