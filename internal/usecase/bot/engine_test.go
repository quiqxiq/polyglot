package bot

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
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

type fakeRetriever struct{}

func (f *fakeRetriever) Retrieve(context.Context, string) ([]knowledge.Entry, error) {
	return nil, nil
}

type fakeLLMConfigRepo struct {
	active *llm.Config
	err    error
}

func (f *fakeLLMConfigRepo) Create(context.Context, *llm.Config) error           { return nil }
func (f *fakeLLMConfigRepo) FindByID(context.Context, uint) (*llm.Config, error) { return f.active, nil }
func (f *fakeLLMConfigRepo) FindActive(context.Context) (*llm.Config, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.active, nil
}
func (f *fakeLLMConfigRepo) FindAll(context.Context) ([]llm.Config, error) { return nil, nil }
func (f *fakeLLMConfigRepo) Update(context.Context, *llm.Config) error       { return nil }
func (f *fakeLLMConfigRepo) SetActive(context.Context, uint) error           { return nil }
func (f *fakeLLMConfigRepo) Delete(context.Context, uint) error              { return nil }

type fakePublisher struct {
	events []string
}

func (f *fakePublisher) PublishEvent(eventType string, _ any) {
	f.events = append(f.events, eventType)
}

// fakeConvRepo implements port.ConversationRepository in memory.
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

func (f *fakeConvRepo) FindActiveConversationByCustomer(_ context.Context, sessionID uint, customerNumber string) (*bot.Conversation, error) {
	for _, c := range f.convs {
		if c.SessionID == sessionID && c.CustomerWANumber == customerNumber && (c.Status == bot.StatusBot || c.Status == bot.StatusEscalation) {
			return c, nil
		}
	}
	return nil, convUC.ErrNotFound
}

func (f *fakeConvRepo) CreateConversation(_ context.Context, conv *bot.Conversation) error {
	conv.ID = f.nextConv
	f.nextConv++
	f.convs[conv.ID] = conv
	return nil
}

func (f *fakeConvRepo) FindConversationByID(_ context.Context, id uint) (*bot.Conversation, error) {
	c, ok := f.convs[id]
	if !ok {
		return nil, convUC.ErrNotFound
	}
	return c, nil
}

func (f *fakeConvRepo) FindConversationByIDWithMessages(_ context.Context, id uint) (*bot.Conversation, error) {
	c, ok := f.convs[id]
	if !ok {
		return nil, convUC.ErrNotFound
	}
	c.Messages = f.msgs[id]
	return c, nil
}

func (f *fakeConvRepo) FindConversationsByStatus(context.Context, bot.ConversationStatus) ([]bot.Conversation, error) {
	return nil, nil
}

func (f *fakeConvRepo) FindConversationsBySessionID(_ context.Context, sessionID uint) ([]bot.Conversation, error) {
	var res []bot.Conversation
	for _, c := range f.convs {
		if c.SessionID == sessionID {
			res = append(res, *c)
		}
	}
	return res, nil
}

func (f *fakeConvRepo) FindAllConversations(context.Context) ([]bot.Conversation, error) {
	return nil, nil
}

func (f *fakeConvRepo) UpdateConversation(_ context.Context, conv *bot.Conversation) error {
	f.convs[conv.ID] = conv
	return nil
}

func (f *fakeConvRepo) CreateMessage(_ context.Context, msg *bot.Message) error {
	msg.ID = f.nextMsg
	f.nextMsg++
	f.msgs[msg.ConversationID] = append(f.msgs[msg.ConversationID], *msg)
	return nil
}

func (f *fakeConvRepo) FindRecentMessages(_ context.Context, conversationID uint, limit int) ([]bot.Message, error) {
	all := f.msgs[conversationID]
	if len(all) <= limit {
		return all, nil
	}
	return all[len(all)-limit:], nil
}

func (f *fakeConvRepo) FindMessagesByConversationID(_ context.Context, conversationID uint) ([]bot.Message, error) {
	return f.msgs[conversationID], nil
}

// fakeChatRepo implements port.ChatRepository with a per-chat bot toggle.
type fakeChatRepo struct {
	botEnabled map[string]bool
}

func newFakeChatRepo() *fakeChatRepo {
	return &fakeChatRepo{botEnabled: make(map[string]bool)}
}

func (f *fakeChatRepo) UpsertChat(context.Context, *bot.WAChat) error                           { return nil }
func (f *fakeChatRepo) UpsertMessage(context.Context, *bot.WAMessage) (bool, error)             { return true, nil }
func (f *fakeChatRepo) UpsertMessagesBatch(context.Context, []*bot.WAMessage) (int, error)     { return len(f.botEnabled), nil }
func (f *fakeChatRepo) IncrementUnread(context.Context, uint, string) error                     { return nil }
func (f *fakeChatRepo) MarkChatRead(context.Context, uint, string) error                        { return nil }
func (f *fakeChatRepo) SetChatUnread(context.Context, uint, string, uint32) error               { return nil }
func (f *fakeChatRepo) ListChats(context.Context, uint, int, int, string) ([]bot.WAChat, error) {
	return nil, nil
}
func (f *fakeChatRepo) ListChatMessages(context.Context, uint, string, int, int) ([]bot.WAMessage, error) {
	return nil, nil
}
func (f *fakeChatRepo) SetChatBotEnabled(_ context.Context, _ uint, chatJID string, enabled bool) error {
	f.botEnabled[chatJID] = enabled
	return nil
}
func (f *fakeChatRepo) IsChatBotEnabled(_ context.Context, _ uint, chatJID string) (bool, error) {
	if enabled, ok := f.botEnabled[chatJID]; ok {
		return enabled, nil
	}
	return true, nil // default aktif
}
func (f *fakeChatRepo) MarkMessagesStatus(context.Context, uint, string, []string, string) error { return nil }
func (f *fakeChatRepo) MergeChatLID(context.Context, uint, string, string) error                          { return nil }

func testBotConfig() Config {
	return Config{
		SystemPrompt:       "Kamu adalah asisten layanan GNET.",
		LLMMaxOutputTokens: 512,
		AllowedTopics:      []string{"internet", "paket", "harga"},
	}
}

func newTestEngine(cache port.CacheStore, gw *fakeGateway, llmRepo *fakeLLMConfigRepo, prov *fakeProvider, convRepo *fakeConvRepo) *Engine {
	return newTestEngineWithChatRepoAndChat(cache, gw, llmRepo, prov, convRepo, newFakeChatRepo(), nil)
}

func newTestEngineWithChatRepo(cache port.CacheStore, gw *fakeGateway, llmRepo *fakeLLMConfigRepo, prov *fakeProvider, convRepo *fakeConvRepo, chatRepo *fakeChatRepo) *Engine {
	return newTestEngineWithChatRepoAndChat(cache, gw, llmRepo, prov, convRepo, chatRepo, nil)
}

func newTestEngineWithChatRepoAndChat(cache port.CacheStore, gw *fakeGateway, llmRepo *fakeLLMConfigRepo, prov *fakeProvider, convRepo *fakeConvRepo, chatRepo *fakeChatRepo, chat port.KnowledgeChat) *Engine {
	svc := convUC.NewConversationService(convRepo)
	return NewEngine(
		testBotConfig(),
		cache,
		gw,
		svc,
		&fakeRetriever{},
		chat,
		llmRepo,
		chatRepo,
		&fakePublisher{},
		func(*llm.Config) (port.LLMProvider, error) { return prov, nil },
	)
}

func TestEngineHappyPath(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "Harga paket mulai Rp150.000/bulan.", tokenIn: 42, tokenOut: 12}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1, Provider: "openai", Model: "gpt-4o", MaxOutputTokens: 256}}
	convRepo := newFakeConvRepo()

	e := newTestEngine(cache, gw, llmRepo, prov, convRepo)
	if err := e.HandleIncomingMessage(context.Background(), 1, "628123456789@s.whatsapp.net", "628123456789", "berapa harga paket internet?"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	if len(gw.sent) != 1 || gw.sent[0] != "Harga paket mulai Rp150.000/bulan." {
		t.Fatalf("expected reply sent, got %v", gw.sent)
	}

	// LLM dipanggil dengan history (user message).
	if len(prov.calls) != 1 {
		t.Fatalf("expected 1 LLM call, got %d", len(prov.calls))
	}

	// Kedua sisi percakapan ter-persist; pesan bot membawa token + llmConfigID.
	var convID uint
	for id := range convRepo.convs {
		convID = id
	}
	msgs := convRepo.msgs[convID]
	if len(msgs) != 2 {
		t.Fatalf("expected 2 persisted messages, got %d", len(msgs))
	}
	botMsg := msgs[1]
	if botMsg.SenderType != bot.SenderBot {
		t.Fatalf("expected second message from bot, got %s", botMsg.SenderType)
	}
	if botMsg.TokenIn != 42 || botMsg.TokenOut != 12 {
		t.Fatalf("tokens not persisted: %+v", botMsg)
	}
	if botMsg.LLMConfigID == nil || *botMsg.LLMConfigID != 1 {
		t.Fatalf("llm config id not persisted: %+v", botMsg.LLMConfigID)
	}

	// Konteks per-chat tersimpan di cache.
	raw, err := cache.Get(context.Background(), "history:conv:1")
	if err != nil || raw == "" {
		t.Fatalf("expected per-conversation history in cache, got %q err=%v", raw, err)
	}
}

// fakeKnowledgeChat is a programmable port.KnowledgeChat for engine tests.
type fakeKnowledgeChat struct {
	result port.KnowledgeChatResult
	err    error
	calls  int
	lastID string
}

func (f *fakeKnowledgeChat) Chat(_ context.Context, msg string, sessionID string) (port.KnowledgeChatResult, error) {
	f.calls++
	f.lastID = sessionID
	if f.err != nil {
		return port.KnowledgeChatResult{}, f.err
	}
	return f.result, nil
}

// TestEngineAnythingLLMPrimary verifies bahwa dengan chat (AnythingLLM)
// terkonfigurasi, jawaban datang dari chat — LLM lokal TIDAK dipanggil.
func TestEngineAnythingLLMPrimary(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "seharusnya tidak dipanggil", tokenIn: 1, tokenOut: 1}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()
	chat := &fakeKnowledgeChat{result: port.KnowledgeChatResult{Content: "Jawaban dari AnythingLLM.", TokenIn: 10, TokenOut: 5}}

	e := newTestEngineWithChatRepoAndChat(cache, gw, llmRepo, prov, convRepo, newFakeChatRepo(), chat)
	if err := e.HandleIncomingMessage(context.Background(), 1, "6281@s.whatsapp.net", "6281", "berapa harga paket?"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	if len(gw.sent) != 1 || gw.sent[0] != "Jawaban dari AnythingLLM." {
		t.Fatalf("expected AnythingLLM reply sent, got %v", gw.sent)
	}
	if len(prov.calls) != 0 {
		t.Fatalf("local LLM should NOT be called when AnythingLLM chat succeeds, got %d calls", len(prov.calls))
	}
	if chat.calls != 1 || chat.lastID != "conv-1" {
		t.Fatalf("expected 1 chat call with session conv-1, got calls=%d lastID=%q", chat.calls, chat.lastID)
	}
}

// TestEngineAnythingLLMFallback verifies bahwa bila chat (AnythingLLM) gagal,
// engine fallback ke LLM lokal proyek.
func TestEngineAnythingLLMFallback(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "Jawaban fallback lokal.", tokenIn: 7, tokenOut: 3}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()
	chat := &fakeKnowledgeChat{err: errors.New("anythingllm down")}

	e := newTestEngineWithChatRepoAndChat(cache, gw, llmRepo, prov, convRepo, newFakeChatRepo(), chat)
	if err := e.HandleIncomingMessage(context.Background(), 1, "6281@s.whatsapp.net", "6281", "berapa harga paket?"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	if len(gw.sent) != 1 || gw.sent[0] != "Jawaban fallback lokal." {
		t.Fatalf("expected fallback reply sent, got %v", gw.sent)
	}
	if len(prov.calls) != 1 {
		t.Fatalf("expected 1 local LLM call after AnythingLLM failure, got %d", len(prov.calls))
	}
}

// TestEngineAnythingLLMEmptyFallback verifies bahwa jawaban kosong dari chat
// juga memicu fallback ke LLM lokal.
func TestEngineAnythingLLMEmptyFallback(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "Jawaban fallback."}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()
	chat := &fakeKnowledgeChat{result: port.KnowledgeChatResult{Content: "   "}}

	e := newTestEngineWithChatRepoAndChat(cache, gw, llmRepo, prov, convRepo, newFakeChatRepo(), chat)
	if err := e.HandleIncomingMessage(context.Background(), 1, "6281@s.whatsapp.net", "6281", "berapa harga paket?"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}
	if len(prov.calls) != 1 {
		t.Fatalf("expected fallback after empty chat reply, got %d local calls", len(prov.calls))
	}
}

func TestEngineEscalationMode(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "balasan"}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()
	svc := convUC.NewConversationService(convRepo)

	conv, err := svc.GetOrCreateConversation(context.Background(), 1, "628123456789")
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.Escalate(context.Background(), conv.ID); err != nil {
		t.Fatal(err)
	}

	e := newTestEngine(cache, gw, llmRepo, prov, convRepo)
	if err := e.HandleIncomingMessage(context.Background(), 1, "628123456789@s.whatsapp.net", "628123456789", "halo"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	if len(gw.sent) != 0 {
		t.Fatalf("no auto-reply expected in escalation mode, got %v", gw.sent)
	}
	if len(prov.calls) != 0 {
		t.Fatalf("LLM must not be called in escalation mode, got %d calls", len(prov.calls))
	}
	// Pesan customer tetap tercatat.
	if len(convRepo.msgs[conv.ID]) != 1 {
		t.Fatalf("expected customer message to be recorded, got %d", len(convRepo.msgs[conv.ID]))
	}
}

func TestEngineNoActiveConfig(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "balasan"}
	llmRepo := &fakeLLMConfigRepo{err: errors.New("no active config")}
	convRepo := newFakeConvRepo()

	e := newTestEngine(cache, gw, llmRepo, prov, convRepo)
	if err := e.HandleIncomingMessage(context.Background(), 1, "628123456789@s.whatsapp.net", "628123456789", "paket apa yang tersedia?"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	if len(gw.sent) != 1 {
		t.Fatalf("expected fallback reply, got %v", gw.sent)
	}
	if len(prov.calls) != 0 {
		t.Fatalf("LLM must not be called without active config")
	}

	var convID uint
	for id, c := range convRepo.convs {
		convID = id
		if c.Status != bot.StatusEscalation {
			t.Fatalf("expected conversation escalated, got %s", c.Status)
		}
	}
	msgs := convRepo.msgs[convID]
	if len(msgs) != 2 {
		t.Fatalf("expected customer + fallback messages, got %d", len(msgs))
	}
}

func TestEngineGetConversationContext(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "oke"}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()
	svc := convUC.NewConversationService(convRepo)

	conv, err := svc.GetOrCreateConversation(context.Background(), 1, "628123456789")
	if err != nil {
		t.Fatal(err)
	}
	llmID := uint(1)
	_, _ = svc.AddMessageWithConfig(context.Background(), conv.ID, "customer", "halo", 0, 0, nil)
	_, _ = svc.AddMessageWithConfig(context.Background(), conv.ID, "bot", "hai", 10, 5, &llmID)
	_, _ = svc.AddMessageWithConfig(context.Background(), conv.ID, "customer", "harga?", 0, 0, nil)
	_, _ = svc.AddMessageWithConfig(context.Background(), conv.ID, "bot", "150rb", 20, 8, &llmID)
	_ = cache.Set(context.Background(), fmt.Sprintf("summary:conv:%d", conv.ID), "Pelanggan tanya harga.", 60)

	e := newTestEngine(cache, gw, llmRepo, prov, convRepo)
	info, err := e.GetConversationContext(context.Background(), conv.ID)
	if err != nil {
		t.Fatalf("GetConversationContext: %v", err)
	}
	if info.Summary != "Pelanggan tanya harga." {
		t.Fatalf("summary missing: %q", info.Summary)
	}
	if info.TotalTokenIn != 30 || info.TotalTokenOut != 13 {
		t.Fatalf("token totals wrong: in=%d out=%d", info.TotalTokenIn, info.TotalTokenOut)
	}
	if info.TotalLLMCalls != 2 {
		t.Fatalf("expected 2 LLM calls, got %d", info.TotalLLMCalls)
	}
	if len(info.RecentMessages) != 4 {
		t.Fatalf("expected 4 recent messages, got %d", len(info.RecentMessages))
	}
	if info.Status != bot.StatusBot || info.ClientPhone != "628123456789" {
		t.Fatalf("unexpected context: %+v", info)
	}

	// Conversation tidak ada → error.
	if _, err := e.GetConversationContext(context.Background(), 999); err == nil {
		t.Fatal("expected error for missing conversation")
	}
}

func TestEngineLLMError(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{err: errors.New("upstream down")}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()

	e := newTestEngine(cache, gw, llmRepo, prov, convRepo)
	if err := e.HandleIncomingMessage(context.Background(), 1, "628123456789@s.whatsapp.net", "628123456789", "paket apa yang tersedia?"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	// LLM gagal → fallback dikirim + percakapan dinaikkan ke escalation.
	if len(gw.sent) != 1 {
		t.Fatalf("expected fallback reply, got %v", gw.sent)
	}
	var convID uint
	for id, c := range convRepo.convs {
		convID = id
		if c.Status != bot.StatusEscalation {
			t.Fatalf("expected conversation escalated after LLM error, got %s", c.Status)
		}
	}
	msgs := convRepo.msgs[convID]
	if len(msgs) != 2 {
		t.Fatalf("expected customer + fallback messages, got %d", len(msgs))
	}
}

func TestEngineOffTopic(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "balasan"}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()

	e := newTestEngine(cache, gw, llmRepo, prov, convRepo)
	if err := e.HandleIncomingMessage(context.Background(), 1, "628123456789@s.whatsapp.net", "628123456789", "ceritakan dongeng tentang kucing"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	if len(gw.sent) != 1 {
		t.Fatalf("expected off-topic reply, got %v", gw.sent)
	}
	if len(prov.calls) != 0 {
		t.Fatalf("LLM must not be called for off-topic messages")
	}
}

func TestEngineChatBotDisabled(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "balasan"}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()
	chatRepo := newFakeChatRepo()
	chatRepo.botEnabled["628123456789@s.whatsapp.net"] = false // agen matikan bot di chat ini

	e := newTestEngineWithChatRepo(cache, gw, llmRepo, prov, convRepo, chatRepo)
	if err := e.HandleIncomingMessage(context.Background(), 1, "628123456789@s.whatsapp.net", "628123456789", "berapa harga paket?"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	// Pesan customer tetap dicatat, tapi tidak ada auto-reply / LLM call.
	if len(gw.sent) != 0 {
		t.Fatalf("no reply expected when chat bot disabled, got %v", gw.sent)
	}
	if len(prov.calls) != 0 {
		t.Fatalf("LLM must not be called when chat bot disabled")
	}
	var convID uint
	for id := range convRepo.convs {
		convID = id
	}
	if len(convRepo.msgs[convID]) != 1 {
		t.Fatalf("expected customer message recorded, got %d", len(convRepo.msgs[convID]))
	}
}

var _ port.WhatsAppGateway = (*fakeGateway)(nil)
var _ port.KnowledgeRetriever = (*fakeRetriever)(nil)
var _ port.LLMConfigRepository = (*fakeLLMConfigRepo)(nil)
var _ port.EventPublisher = (*fakePublisher)(nil)
var _ port.ChatRepository = (*fakeChatRepo)(nil)
var _ port.ConversationRepository = (*fakeConvRepo)(nil)
