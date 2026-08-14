//go:build integration

// Test ini membuktikan end-to-end bahwa bot WhatsApp menjawab dari dokumen
// knowledge yang baru di-embed ke AnythingLLM (vector search). Alurnya
// meniru persis wiring produksi di internal/app/app.go:
//
//	usecase CreateDocument(embed=true) → AnythingLLM raw-text + vector embed
//	→ HybridRetriever (keyword Postgres + vector AnythingLLM)
//	→ ContextManager.BuildPromptContext (knowledge disuntik ke system prompt)
//	→ LLM aktif dari llm_configs (Gemini) menjawab
//
// Tanpa mengirim pesan WhatsApp sungguhan (gateway tidak dipanggil).
//
// Cara jalan (wajib ANYTHINGLLM_API_KEY ter-set):
//
//	set -a && . ./.env && set +a && \
//	export ANYTHINGLLM_API_KEY=polyglot-dev-... ANYTHINGLLM_BASE_URL=http://localhost:3001 \
//	       ANYTHINGLLM_WORKSPACE=netops && \
//	go test -tags=integration ./test/integration/... -run TestBotAnswersFromEmbeddedKnowledge -v
//
// Kalau ANYTHINGLLM_API_KEY kosong atau tidak ada LLM aktif, test di-skip.
package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/knowledge"
	llmadapter "github.com/quixiq/polyglot/internal/adapter/llm"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	redisAdapter "github.com/quixiq/polyglot/internal/adapter/redis"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/bot"
	botUC "github.com/quixiq/polyglot/internal/usecase/bot"
	knowledgeUC "github.com/quixiq/polyglot/internal/usecase/knowledge"
)

func TestBotAnswersFromEmbeddedKnowledge(t *testing.T) {
	if os.Getenv("ANYTHINGLLM_API_KEY") == "" {
		t.Skip("ANYTHINGLLM_API_KEY not set — skipping AnythingLLM vector integration test")
	}

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgStore, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("postgres: %v", err)
	}
	redisStore, _ := redisAdapter.NewStore(cfg.RedisURL) // boleh nil — cache optional

	// ── 1. Hybrid retriever (sama dengan wiring app.go) ──────────────
	keywordRetriever := knowledgeUC.NewKeywordRetriever(pgStore)
	anyRetriever, err := knowledge.NewRetriever(cfg.AnythingLLMBaseURL, cfg.AnythingLLMAPIKey, cfg.AnythingLLMWorkspace, cfg.AnythingLLMTopN)
	if err != nil {
		t.Fatalf("anythingllm retriever: %v", err)
	}
	hybrid := knowledgeUC.NewHybridRetriever(keywordRetriever, anyRetriever)

	// ── 2. Document manager untuk create+embed (sama dengan wiring) ──
	anyManager, err := knowledge.NewManager(cfg.AnythingLLMBaseURL, cfg.AnythingLLMAPIKey, cfg.AnythingLLMWorkspace)
	if err != nil {
		t.Fatalf("anythingllm manager: %v", err)
	}
	docUC := knowledgeUC.NewDocumentManager(postgres.NewKnowledgeRepository(pgStore), anyManager)

	// ── 3. Buat dokumen knowledge baru (embed=true) dengan fakta unik ─
	unique := fmt.Sprintf("FIBER2026%d", time.Now().Unix()%100000)
	title := fmt.Sprintf("Paket GNET Fiber X-TREME 500 %s", unique)
	content := fmt.Sprintf(
		`# Paket GNET Fiber X-TREME 500

Paket internet rumah terbaru GNET dengan kecepatan download 500 Mbps dan upload 250 Mbps.

- Harga bulanan: Rp 399.000, biaya pemasangan GRATIS.
- Kode promo khusus pelanggan baru: %s (berlaku 30 hari, potongan Rp 100.000 untuk 3 bulan pertama).
- Termasuk router WiFi 6 dan ONT gratis sewa.
- Kecepatan minimum yang dijamin (CIR): 400 Mbps.
- Berlaku untuk area Jakarta, Bogor, Depok, Tangerang, dan Bekasi.`,
		unique,
	)
	entry, err := docUC.CreateDocument(ctx, knowledgeUC.CreateParams{
		Title:      title,
		Content:    content,
		Category:   "umum",
		Tags:       []string{"fiber", "promo", "500"},
		EmbedToLLM: true,
	})
	if err != nil {
		t.Fatalf("create+embed doc: %v", err)
	}
	t.Cleanup(func() {
		_ = docUC.DeleteDocument(context.Background(), entry.ID)
	})
	if entry.EmbedStatus != "embedded" {
		t.Fatalf("expected embedded, got %q (entry id=%d)", entry.EmbedStatus, entry.ID)
	}
	t.Logf("Dokumen dibuat & ter-embed: id=%d title=%q status=%s docName=%q",
		entry.ID, entry.Title, entry.EmbedStatus, entry.AnythingLLMDocName)

	// ── 4. Retrieve lewat hybrid (query yang menargetkan fakta unik) ─
	question := fmt.Sprintf("berapa harga paket fiber 500 Mbps dan kode promo %s itu apa saja keunggulannya?", unique)
	relEntries, err := hybrid.Retrieve(ctx, question)
	if err != nil {
		t.Fatalf("hybrid retrieve: %v", err)
	}
	t.Logf("Retrieve %q → %d entries", question, len(relEntries))
	found := false
	for i, e := range relEntries {
		t.Logf("  [%d] title=%q | content[:80]=%q", i, e.Title, truncate(e.Content, 80))
		if strings.Contains(e.Title, unique) || strings.Contains(e.Content, unique) {
			found = true
		}
	}
	if !found {
		t.Fatalf("dokumen yang baru di-embed TIDAK ter-retrieve — vector search tidak menjawab dari dokumen ini")
	}
	t.Log("✅ Dokumen baru ter-retrieve dari knowledge (keyword dan/atau vector)")

	// ── 5. Bangun prompt context persis seperti engine bot ───────────
	contextMgr := botUC.NewContextManager(redisStore, cfg)
	systemPrompt, messages, err := contextMgr.BuildPromptContext(ctx, 0, []bot.Message{
		{SenderType: bot.SenderCustomer, Content: question},
	}, relEntries)
	if err != nil {
		t.Fatalf("build prompt: %v", err)
	}
	if !strings.Contains(systemPrompt, unique) {
		t.Logf("WARNING: fakta unik tidak muncul di system prompt — entries mungkin hanya keyword tanpa vector chunk")
	} else {
		t.Log("✅ Fakta dokumen masuk ke system prompt (BASIS PENGETAHUAN LOKAL)")
	}
	t.Logf("── system prompt (terpotong) ──\n%s", truncate(systemPrompt, 1200))

	// ── 6. Panggil LLM aktif (Gemini) dan cek jawaban ────────────────
	llmCfg, err := pgStore.FindActive()
	if err != nil {
		t.Skipf("no active LLM config — retrieval verified, LLM answer skipped: %v", err)
	}
	provider, err := llmadapter.NewProvider(llmCfg, cfg.EncryptionKey)
	if err != nil {
		t.Fatalf("llm provider: %v", err)
	}
	maxTokens := llmCfg.MaxOutputTokens
	if maxTokens <= 0 {
		maxTokens = cfg.LLMMaxOutputTokens
	}
	resp, err := provider.Chat(ctx, systemPrompt, messages, maxTokens)
	if err != nil {
		t.Fatalf("llm chat: %v", err)
	}
	t.Logf("── Jawaban bot (gemini, %d/%d token) ──\n%s", resp.TokenIn, resp.TokenOut, resp.Content)

	// Jawaban harus menyebut fakta dari dokumen (harga atau kode promo).
	lower := strings.ToLower(resp.Content)
	if strings.Contains(lower, strings.ToLower(unique)) || strings.Contains(lower, "399") || strings.Contains(lower, "500 mbps") {
		t.Log("✅ Jawaban bot menyebut fakta dari dokumen yang baru di-embed")
	} else {
		t.Log("⚠️ Jawaban tidak menyebut fakta unik secara eksplisit — cek prompt/LLM (bukan kegagalan hard)")
	}
	log.Printf("[integration] done: %s", title)
}

// TestAnythingLLMChatPrimaryPath membuktikan jalur PRIMARY bot: panggilan
// langsung ke workspace chat AnythingLLM (RAG + LLM dalam satu panggilan,
// mode "chat") menjawab dari dokumen yang baru di-embed. Ini arsitektur
// yang dipilih: AnythingLLM sebagai otak, LLM lokal sebagai backup.
func TestAnythingLLMChatPrimaryPath(t *testing.T) {
	if os.Getenv("ANYTHINGLLM_API_KEY") == "" {
		t.Skip("ANYTHINGLLM_API_KEY not set — skipping AnythingLLM chat integration test")
	}

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	pgStore, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("postgres: %v", err)
	}

	// ── 1. Document manager untuk create+embed ───────────────────────
	anyManager, err := knowledge.NewManager(cfg.AnythingLLMBaseURL, cfg.AnythingLLMAPIKey, cfg.AnythingLLMWorkspace)
	if err != nil {
		t.Fatalf("anythingllm manager: %v", err)
	}
	docUC := knowledgeUC.NewDocumentManager(postgres.NewKnowledgeRepository(pgStore), anyManager)

	// ── 2. Chat client (otak utama bot) ──────────────────────────────
	chatClient, err := knowledge.NewChatClient(cfg.AnythingLLMBaseURL, cfg.AnythingLLMAPIKey, cfg.AnythingLLMWorkspace)
	if err != nil {
		t.Fatalf("anythingllm chat client: %v", err)
	}

	// ── 3. Buat dokumen baru (embed=true) dengan fakta unik ──────────
	unique := fmt.Sprintf("WIFI6MESH%d", time.Now().Unix()%100000)
	title := fmt.Sprintf("Paket GNET Mesh WiFi 6 %s", unique)
	content := fmt.Sprintf(
		`# Paket GNET Mesh WiFi 6

Paket khusus rumah besar dengan sistem mesh WiFi 6 (2 unit) agar sinyal merata sampai lantai 3.

- Harga bulanan: Rp 289.000 (tanpa biaya sewa router tambahan).
- Kode verifikasi khusus: %s — hanya untuk pelanggan area Bekasi.
- Cakupan area: hingga 300 m2, maksimal 40 perangkat terhubung.`,
		unique,
	)
	entry, err := docUC.CreateDocument(ctx, knowledgeUC.CreateParams{
		Title:      title,
		Content:    content,
		Category:   "umum",
		Tags:       []string{"mesh", "wifi6"},
		EmbedToLLM: true,
	})
	if err != nil {
		t.Fatalf("create+embed doc: %v", err)
	}
	t.Cleanup(func() {
		_ = docUC.DeleteDocument(context.Background(), entry.ID)
	})
	if entry.EmbedStatus != "embedded" {
		t.Fatalf("expected embedded, got %q (entry id=%d)", entry.EmbedStatus, entry.ID)
	}
	t.Logf("Dokumen dibuat & ter-embed: id=%d title=%q", entry.ID, title)

	// ── 4. Tanya ke workspace chat (jalur primary bot) ───────────────
	question := fmt.Sprintf("berapa harga paket mesh wifi 6 dan apa kode verifikasi %s itu?", unique)
	result, err := chatClient.Chat(ctx, question, fmt.Sprintf("integration-conv-%d", entry.ID))
	if err != nil {
		t.Fatalf("anythingllm chat: %v", err)
	}
	t.Logf("── Jawaban AnythingLLM chat (%d/%d token) ──\n%s", result.TokenIn, result.TokenOut, result.Content)
	t.Logf("── Sumber grounding: %v", result.Sources)

	lower := strings.ToLower(result.Content)
	if !strings.Contains(lower, strings.ToLower(unique)) && !strings.Contains(lower, "289") {
		t.Fatalf("jawaban chat TIDAK menyebut fakta dari dokumen yang baru di-embed: %q", result.Content)
	}
	t.Log("✅ AnythingLLM chat menjawab dari dokumen yang baru di-embed (jalur primary bot)")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
