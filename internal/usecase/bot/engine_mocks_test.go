package bot

import (
	"context"

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

func (f *fakeLLMConfigRepo) Create(context.Context, *llm.Config) error { return nil }
func (f *fakeLLMConfigRepo) FindByID(context.Context, uint) (*llm.Config, error) {
	return f.active, nil
}
func (f *fakeLLMConfigRepo) FindActive(context.Context) (*llm.Config, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.active, nil
}
func (f *fakeLLMConfigRepo) FindAll(context.Context) ([]llm.Config, error) { return nil, nil }
func (f *fakeLLMConfigRepo) Update(context.Context, *llm.Config) error     { return nil }
func (f *fakeLLMConfigRepo) SetActive(context.Context, uint) error         { return nil }
func (f *fakeLLMConfigRepo) Delete(context.Context, uint) error            { return nil }

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

func (f *fakeChatRepo) UpsertChat(context.Context, *bot.WAChat) error { return nil }
func (f *fakeChatRepo) UpsertMessage(context.Context, *bot.WAMessage) (bool, error) {
	return true, nil
}
func (f *fakeChatRepo) UpsertMessagesBatch(context.Context, []*bot.WAMessage) (int, error) {
	return len(f.botEnabled), nil
}
func (f *fakeChatRepo) IncrementUnread(context.Context, uint, string) error { return nil }
func (f *fakeChatRepo) MarkChatRead(context.Context, uint, string) error    { return nil }
func (f *fakeChatRepo) SetChatUnread(context.Context, uint, string, uint32) error {
	return nil
}
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
func (f *fakeChatRepo) MarkMessagesStatus(context.Context, uint, string, []string, string) error {
	return nil
}
func (f *fakeChatRepo) MergeChatLID(context.Context, uint, string, string) error { return nil }

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

var _ port.WhatsAppGateway = (*fakeGateway)(nil)
var _ port.KnowledgeRetriever = (*fakeRetriever)(nil)
var _ port.LLMConfigRepository = (*fakeLLMConfigRepo)(nil)
var _ port.EventPublisher = (*fakePublisher)(nil)
var _ port.ChatRepository = (*fakeChatRepo)(nil)
var _ port.ConversationRepository = (*fakeConvRepo)(nil)
