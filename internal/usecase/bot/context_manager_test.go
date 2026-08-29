package bot

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
)

// fakeCache is a minimal in-memory CacheStore.
type fakeCache struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeCache() *fakeCache {
	return &fakeCache{data: make(map[string]string)}
}

func (f *fakeCache) Get(_ context.Context, key string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if v, ok := f.data[key]; ok {
		return v, nil
	}
	return "", errors.New("cache miss")
}

func (f *fakeCache) Set(_ context.Context, key, value string, _ int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.data[key] = value
	return nil
}

func (f *fakeCache) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.data, key)
	return nil
}

func (f *fakeCache) Close() error { return nil }

type fakeProvider struct {
	reply             string
	tokenIn, tokenOut int
	err               error
	calls             []string
}

func (f *fakeProvider) Chat(_ context.Context, systemPrompt string, messages []llm.ChatMessage, _ int) (*llm.ChatResponse, error) {
	contents := make([]string, 0, len(messages))
	for _, m := range messages {
		contents = append(contents, m.Role+":"+m.Content)
	}
	f.calls = append(f.calls, systemPrompt+" || "+strings.Join(contents, "; "))
	if f.err != nil {
		return nil, f.err
	}
	return &llm.ChatResponse{Content: f.reply, TokenIn: f.tokenIn, TokenOut: f.tokenOut}, nil
}

func (f *fakeProvider) ChatWithTools(ctx context.Context, systemPrompt string, messages []llm.ChatMessage, _ []llm.Tool, maxTokens int) (*llm.ChatResponse, error) {
	return f.Chat(ctx, systemPrompt, messages, maxTokens)
}

func (f *fakeProvider) TestConnection(context.Context) error { return nil }

func TestBuildPromptContextRolesAndSummary(t *testing.T) {
	cache := newFakeCache()
	cm := NewContextManager(cache)

	ctx := context.Background()

	// Siapkan summary per-conversation.
	if err := cache.Set(ctx, "summary:conv:5", "Pelanggan menanyakan paket.", 60); err != nil {
		t.Fatal(err)
	}

	history := []bot.Message{
		{ID: 1, ConversationID: 5, SenderType: bot.SenderCustomer, Content: "halo"},
		{ID: 2, ConversationID: 5, SenderType: bot.SenderBot, Content: "Halo, ada yang bisa dibantu?"},
		{ID: 3, ConversationID: 5, SenderType: bot.SenderAgent, Content: "Ini dari admin."},
	}

	compositePrompt := "Kamu adalah asisten layanan GNET.\n\n### SKILL: Paket\nPaket mulai 150rb."
	systemPrompt, messages, err := cm.BuildPromptContext(ctx, 5, history, compositePrompt)
	if err != nil {
		t.Fatalf("BuildPromptContext: %v", err)
	}

	if !strings.Contains(systemPrompt, "Kamu adalah asisten layanan GNET.") {
		t.Fatal("base system prompt missing")
	}
	if !strings.Contains(systemPrompt, "Paket mulai 150rb.") {
		t.Fatal("skill entries missing from prompt")
	}
	if !strings.Contains(systemPrompt, "Pelanggan menanyakan paket.") {
		t.Fatal("per-conversation summary missing from prompt")
	}

	if len(messages) != 3 {
		t.Fatalf("expected 3 chat messages, got %d", len(messages))
	}
	if messages[0].Role != "user" || messages[1].Role != "assistant" || messages[2].Role != "assistant" {
		t.Fatalf("role mapping wrong: %+v", messages)
	}

	// Summary milik conv lain tidak boleh bocor.
	other, _, err := cm.BuildPromptContext(ctx, 99, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(other, "Pelanggan menanyakan paket.") {
		t.Fatal("summary leaked across conversations")
	}
}

func TestSaveMessageToSessionAppendsAndTrims(t *testing.T) {
	cache := newFakeCache()
	cm := NewContextManager(cache)
	ctx := context.Background()

	// Setiap save menyimpan 2 pesan (user + bot). 25 turn = 50 pesan >
	// maxCachedTurns, jadi window harus ter-trim ke 40 pesan = 20 turn.
	for i := 0; i < 25; i++ {
		if err := cm.SaveMessageToSession(ctx, 5, "user", "bot"); err != nil {
			t.Fatalf("SaveMessageToSession: %v", err)
		}
	}

	raw, _ := cache.Get(ctx, "history:conv:5")
	turns := strings.Count(raw, `"role":"user"`)
	if turns != maxCachedTurns/2 {
		t.Fatalf("expected window trimmed to %d turns, got %d", maxCachedTurns/2, turns)
	}

	// Tambah terus — window harus tetap pada batasnya.
	for i := 0; i < 30; i++ {
		_ = cm.SaveMessageToSession(ctx, 5, "user", "bot")
	}
	raw, _ = cache.Get(ctx, "history:conv:5")
	turns = strings.Count(raw, `"role":"user"`)
	if turns != maxCachedTurns/2 {
		t.Fatalf("expected window capped at %d turns, got %d", maxCachedTurns/2, turns)
	}
}

func TestSummarizeSessionIfLong(t *testing.T) {
	cache := newFakeCache()
	cm := NewContextManager(cache)
	ctx := context.Background()
	prov := &fakeProvider{reply: "Ringkasan percakapan."}

	// Ambang dihitung per PESAN (user+bot = 2 per turn). 8 turn = 16 pesan <
	// summarizeThreshold — tidak boleh memanggil LLM.
	for i := 0; i < summarizeThreshold/2-2; i++ {
		_ = cm.SaveMessageToSession(ctx, 5, "user", "bot")
	}
	if err := cm.SummarizeSessionIfLong(ctx, 5, prov); err != nil {
		t.Fatalf("SummarizeSessionIfLong (short): %v", err)
	}
	if len(prov.calls) != 0 {
		t.Fatalf("LLM should not be called under threshold, got %d calls", len(prov.calls))
	}

	// Tambah hingga 12 turn = 24 pesan >= summarizeThreshold — ringkas & bersihkan.
	for i := 0; i < 4; i++ {
		_ = cm.SaveMessageToSession(ctx, 5, "user", "bot")
	}
	if err := cm.SummarizeSessionIfLong(ctx, 5, prov); err != nil {
		t.Fatalf("SummarizeSessionIfLong (long): %v", err)
	}
	if len(prov.calls) != 1 {
		t.Fatalf("expected 1 summarization call, got %d", len(prov.calls))
	}
	summary, _ := cache.Get(ctx, "summary:conv:5")
	if summary != "Ringkasan percakapan." {
		t.Fatalf("unexpected summary: %q", summary)
	}
	if _, err := cache.Get(ctx, "history:conv:5"); err == nil {
		t.Fatal("history should be cleared after summarization")
	}

	// Provider nil — aman.
	if err := cm.SummarizeSessionIfLong(ctx, 5, nil); err != nil {
		t.Fatalf("nil provider: %v", err)
	}
}

var _ port.CacheStore = (*fakeCache)(nil)
var _ port.LLMProvider = (*fakeProvider)(nil)
