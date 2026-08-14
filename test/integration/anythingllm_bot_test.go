//go:build integration

package integration

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	knowledgeadapter "github.com/quixiq/polyglot/internal/adapter/knowledge"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
	botUC "github.com/quixiq/polyglot/internal/usecase/bot"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
)

// fakeGateway implements port.WhatsAppGateway and records sent messages.
type fakeGateway struct {
	sent []string
}

func (f *fakeGateway) Connect(*bot.WASession) error { return nil }
func (f *fakeGateway) Disconnect(uint) error        { return nil }
func (f *fakeGateway) Logout(uint) error            { return nil }
func (f *fakeGateway) Purge(uint) error             { return nil }
func (f *fakeGateway) Reconnect(uint) error         { return nil }
func (f *fakeGateway) SendMessage(_ uint, _ string, content string) error {
	f.sent = append(f.sent, content)
	return nil
}
func (f *fakeGateway) SendDocument(context.Context, uint, string, []byte, string, string, string) error {
	return nil
}
func (f *fakeGateway) SendImage(context.Context, uint, string, []byte, string, string) error {
	return nil
}
func (f *fakeGateway) GetStatus(uint) (string, error)              { return "online", nil }
func (f *fakeGateway) GetQRCode(uint) (string, error)              { return "", nil }
func (f *fakeGateway) GetPairingCode(uint, string) (string, error) { return "", nil }
func (f *fakeGateway) RestoreAllSessions([]bot.WASession) error    { return nil }

var _ port.WhatsAppGateway = (*fakeGateway)(nil)

// fakeCache implements port.CacheStore in memory.
type fakeCache struct {
	data map[string]string
}

func newFakeCache() *fakeCache { return &fakeCache{data: map[string]string{}} }

func (c *fakeCache) Get(_ context.Context, key string) (string, error) { return c.data[key], nil }
func (c *fakeCache) Set(_ context.Context, key string, value string, _ int) error {
	c.data[key] = value
	return nil
}
func (c *fakeCache) Delete(_ context.Context, key string) error {
	delete(c.data, key)
	return nil
}

var _ port.CacheStore = (*fakeCache)(nil)

// fakeChatRepo implements port.ChatRepository with per-chat bot toggle aktif.
type fakeChatRepo struct{}

func (f *fakeChatRepo) UpsertChat(*bot.WAChat) error                           { return nil }
func (f *fakeChatRepo) UpsertMessage(*bot.WAMessage) (bool, error)             { return true, nil }
func (f *fakeChatRepo) UpsertMessagesBatch(msgs []*bot.WAMessage) (int, error) { return len(msgs), nil }
func (f *fakeChatRepo) IncrementUnread(uint, string) error                     { return nil }
func (f *fakeChatRepo) MarkChatRead(uint, string) error                        { return nil }
func (f *fakeChatRepo) SetChatUnread(uint, string, uint32) error               { return nil }
func (f *fakeChatRepo) ListChats(uint, int, int, string) ([]bot.WAChat, error) { return nil, nil }
func (f *fakeChatRepo) ListChatMessages(uint, string, int, int) ([]bot.WAMessage, error) {
	return nil, nil
}
func (f *fakeChatRepo) SetChatBotEnabled(uint, string, bool) error { return nil }
func (f *fakeChatRepo) IsChatBotEnabled(uint, string) (bool, error) {
	return true, nil // default aktif
}
func (f *fakeChatRepo) MarkMessagesStatus(uint, string, []string, string) error { return nil }
func (f *fakeChatRepo) MergeChatLID(uint, string, string) error                 { return nil }

var _ port.ChatRepository = (*fakeChatRepo)(nil)

// fakePublisher implements port.EventPublisher.
type fakePublisher struct{}

func (f *fakePublisher) PublishEvent(string, any) {}

var _ port.EventPublisher = (*fakePublisher)(nil)

// fakeLLMConfigRepo returns an active config so the engine proceeds to the LLM.
type fakeLLMConfigRepo struct{}

func (f *fakeLLMConfigRepo) Create(*llm.LLMConfig) error           { return nil }
func (f *fakeLLMConfigRepo) FindByID(uint) (*llm.LLMConfig, error) { return &llm.LLMConfig{ID: 1}, nil }
func (f *fakeLLMConfigRepo) FindActive() (*llm.LLMConfig, error)   { return &llm.LLMConfig{ID: 1}, nil }
func (f *fakeLLMConfigRepo) FindAll() ([]llm.LLMConfig, error)     { return nil, nil }
func (f *fakeLLMConfigRepo) Update(*llm.LLMConfig) error           { return nil }
func (f *fakeLLMConfigRepo) SetActive(uint) error                  { return nil }
func (f *fakeLLMConfigRepo) Delete(uint) error                     { return nil }

var _ port.LLMConfigRepository = (*fakeLLMConfigRepo)(nil)

// fakeProvider captures the systemPrompt that the real retriever produced, and
// answers from the knowledge found there — persis jalur yang dipakai engine.
type fakeProvider struct {
	systemPrompt string
	reply        string
}

func (p *fakeProvider) Chat(_ context.Context, systemPrompt string, _ []llm.ChatMessage, _ int) (*llm.ChatResponse, error) {
	p.systemPrompt = systemPrompt
	return &llm.ChatResponse{Content: p.reply, TokenIn: 10, TokenOut: 5}, nil
}
func (p *fakeProvider) TestConnection(context.Context) error { return nil }

var _ port.LLMProvider = (*fakeProvider)(nil)

// fakeConvRepo implements conversation.ConversationRepository in memory.
type fakeConvRepo struct {
	convs    map[uint]*bot.Conversation
	msgs     map[uint][]bot.Message
	nextConv uint
	nextMsg  uint
}

func newFakeConvRepo() *fakeConvRepo {
	return &fakeConvRepo{
		convs:    make(map[uint]*bot.Conversation),
		msgs:     make(map[uint][]bot.Message),
		nextConv: 1,
		nextMsg:  1,
	}
}

func (f *fakeConvRepo) FindActiveConversationByCustomer(sessionID uint, customerNumber string) (*bot.Conversation, error) {
	for _, c := range f.convs {
		if c.SessionID == sessionID && c.CustomerWANumber == customerNumber && (c.Status == bot.StatusBot || c.Status == bot.StatusEscalation) {
			return c, nil
		}
	}
	return nil, convUC.ErrNotFound
}

func (f *fakeConvRepo) CreateConversation(conv *bot.Conversation) error {
	conv.ID = f.nextConv
	f.nextConv++
	f.convs[conv.ID] = conv
	return nil
}

func (f *fakeConvRepo) FindConversationByID(id uint) (*bot.Conversation, error) {
	c, ok := f.convs[id]
	if !ok {
		return nil, convUC.ErrNotFound
	}
	return c, nil
}

func (f *fakeConvRepo) FindConversationByIDWithMessages(id uint) (*bot.Conversation, error) {
	c, ok := f.convs[id]
	if !ok {
		return nil, convUC.ErrNotFound
	}
	c.Messages = f.msgs[id]
	return c, nil
}

func (f *fakeConvRepo) FindConversationsByStatus(bot.ConversationStatus) ([]bot.Conversation, error) {
	return nil, nil
}

func (f *fakeConvRepo) FindConversationsBySessionID(sessionID uint) ([]bot.Conversation, error) {
	var res []bot.Conversation
	for _, c := range f.convs {
		if c.SessionID == sessionID {
			res = append(res, *c)
		}
	}
	return res, nil
}

func (f *fakeConvRepo) FindAllConversations() ([]bot.Conversation, error) { return nil, nil }

func (f *fakeConvRepo) UpdateConversation(conv *bot.Conversation) error {
	f.convs[conv.ID] = conv
	return nil
}

func (f *fakeConvRepo) CreateMessage(msg *bot.Message) error {
	msg.ID = f.nextMsg
	f.nextMsg++
	f.msgs[msg.ConversationID] = append(f.msgs[msg.ConversationID], *msg)
	return nil
}

func (f *fakeConvRepo) FindRecentMessages(conversationID uint, limit int) ([]bot.Message, error) {
	all := f.msgs[conversationID]
	if len(all) <= limit {
		return all, nil
	}
	return all[len(all)-limit:], nil
}

var _ convUC.ConversationRepository = (*fakeConvRepo)(nil)

// TestAnythingLLMKnowledgeViaBot memverifikasi rantai end-to-end yang dipakai
// bot sungguhan: message WhatsApp masuk → retriever mengambil chunk dari
// AnythingLLM live (workspace netops) → chunk masuk system prompt ("BASIS
// PENGETAHUAN LOKAL") → provider dipanggil → reply dikirim ke gateway.
//
// Prasyarat (integration): AnythingLLM jalan di ANYTHINGLLM_BASE_URL dengan
// workspace ANYTHINGLLM_WORKSPACE yang sudah berisi dokumen, dan
// ANYTHINGLLM_API_KEY valid. Tanpa key, test di-skip.
func TestAnythingLLMKnowledgeViaBot(t *testing.T) {
	baseURL := os.Getenv("ANYTHINGLLM_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3001"
	}
	apiKey := os.Getenv("ANYTHINGLLM_API_KEY")
	if apiKey == "" {
		t.Skip("ANYTHINGLLM_API_KEY tidak di-set — skip integration test bot + AnythingLLM")
	}
	workspace := os.Getenv("ANYTHINGLLM_WORKSPACE")
	if workspace == "" {
		workspace = "netops"
	}

	retriever, err := knowledgeadapter.NewRetriever(baseURL, apiKey, workspace, 6)
	if err != nil {
		t.Fatalf("NewRetriever: %v", err)
	}

	cfg := config.Load()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "Paket GNET 20 Mbps seharga Rp250.000/bulan."}
	convRepo := newFakeConvRepo()
	svc := convUC.NewConversationService(convRepo)

	// chat (AnythingLLM workspace chat) sengaja nil supaya test ini
	// memverifikasi jalur FALLBACK: retriever + LLM lokal proyek.
	engine := botUC.NewEngine(
		cfg,
		newFakeCache(),
		gw,
		svc,
		retriever,
		nil,
		&fakeLLMConfigRepo{},
		&fakeChatRepo{},
		&fakePublisher{},
		func(*llm.LLMConfig) (port.LLMProvider, error) { return prov, nil },
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = engine.HandleIncomingMessage(ctx, 1, "628123456789@s.whatsapp.net", "628123456789", "berapa harga paket 20 mbps?")
	if err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	// 1) Chunk knowledge dari AnythingLLM benar-benar sampai ke prompt LLM.
	if !strings.Contains(prov.systemPrompt, "BASIS PENGETAHUAN LOKAL") {
		t.Fatalf("system prompt tidak mengandung section knowledge: %q", prov.systemPrompt[:min(len(prov.systemPrompt), 300)])
	}
	if !strings.Contains(prov.systemPrompt, "paket-internet.md") {
		t.Fatalf("chunk dari dokumen paket-internet.md tidak ditemukan di prompt: %q", prov.systemPrompt[:min(len(prov.systemPrompt), 500)])
	}
	if !strings.Contains(prov.systemPrompt, "Rp250.000") {
		t.Fatalf("isi dokumen (harga) tidak sampai ke prompt: %q", prov.systemPrompt[:min(len(prov.systemPrompt), 500)])
	}

	// 2) Reply dikirim ke gateway WhatsApp.
	if len(gw.sent) != 1 || gw.sent[0] != prov.reply {
		t.Fatalf("expected 1 reply sent via gateway, got %v", gw.sent)
	}

	// 3) Kedua sisi percakapan ter-persist.
	var convID uint
	for id := range convRepo.convs {
		convID = id
	}
	if len(convRepo.msgs[convID]) != 2 {
		t.Fatalf("expected 2 persisted messages (customer + bot), got %d", len(convRepo.msgs[convID]))
	}
}
