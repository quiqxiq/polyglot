package conversation

import (
	"context"
	"errors"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/port"
)

type fakeRepo struct {
	convs     map[uint]*bot.Conversation
	msgs      map[uint][]bot.Message
	nextConv  uint
	nextMsg   uint
	createErr error
}

var _ port.ConversationRepository = (*fakeRepo)(nil)

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		convs:    make(map[uint]*bot.Conversation),
		msgs:     make(map[uint][]bot.Message),
		nextConv: 1,
		nextMsg:  1,
	}
}

func (f *fakeRepo) FindActiveConversationByCustomer(ctx context.Context, sessionID uint, customerNumber string) (*bot.Conversation, error) {
	for _, c := range f.convs {
		if c.SessionID == sessionID && c.CustomerWANumber == customerNumber && (c.Status == bot.StatusBot || c.Status == bot.StatusEscalation) {
			return c, nil
		}
	}
	return nil, bot.ErrConversationNotFound
}

func (f *fakeRepo) CreateConversation(ctx context.Context, conv *bot.Conversation) error {
	conv.ID = f.nextConv
	f.nextConv++
	f.convs[conv.ID] = conv
	return nil
}

func (f *fakeRepo) FindConversationByID(ctx context.Context, id uint) (*bot.Conversation, error) {
	c, ok := f.convs[id]
	if !ok {
		return nil, bot.ErrConversationNotFound
	}
	return c, nil
}

func (f *fakeRepo) FindConversationByIDWithMessages(ctx context.Context, id uint) (*bot.Conversation, error) {
	c, ok := f.convs[id]
	if !ok {
		return nil, bot.ErrConversationNotFound
	}
	c.Messages = f.msgs[id]
	return c, nil
}

func (f *fakeRepo) FindConversationsByStatus(ctx context.Context, status bot.ConversationStatus) ([]bot.Conversation, error) {
	var res []bot.Conversation
	for _, c := range f.convs {
		if c.Status == status {
			res = append(res, *c)
		}
	}
	return res, nil
}

func (f *fakeRepo) FindConversationsBySessionID(ctx context.Context, sessionID uint) ([]bot.Conversation, error) {
	var res []bot.Conversation
	for _, c := range f.convs {
		if c.SessionID == sessionID {
			res = append(res, *c)
		}
	}
	return res, nil
}

func (f *fakeRepo) FindAllConversations(ctx context.Context) ([]bot.Conversation, error) {
	var res []bot.Conversation
	for _, c := range f.convs {
		res = append(res, *c)
	}
	return res, nil
}

func (f *fakeRepo) UpdateConversation(ctx context.Context, conv *bot.Conversation) error {
	f.convs[conv.ID] = conv
	return nil
}

func (f *fakeRepo) CreateMessage(ctx context.Context, msg *bot.Message) error {
	if f.createErr != nil {
		return f.createErr
	}
	msg.ID = f.nextMsg
	f.nextMsg++
	f.msgs[msg.ConversationID] = append(f.msgs[msg.ConversationID], *msg)
	return nil
}

func (f *fakeRepo) FindMessagesByConversationID(ctx context.Context, conversationID uint) ([]bot.Message, error) {
	return f.msgs[conversationID], nil
}

func (f *fakeRepo) FindRecentMessages(ctx context.Context, conversationID uint, limit int) ([]bot.Message, error) {
	all := f.msgs[conversationID]
	if len(all) <= limit {
		return all, nil
	}
	return all[len(all)-limit:], nil
}

func TestAddMessageWithConfigPersists(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewConversationUseCase(repo)

	conv, err := svc.GetOrCreateConversation(ctx, 1, "628123456789")
	if err != nil {
		t.Fatalf("GetOrCreateConversation: %v", err)
	}
	if conv.ID == 0 {
		t.Fatal("expected conversation to get an ID from repo")
	}

	msg, err := svc.AddMessageWithConfig(ctx, conv.ID, "customer", "berapa harga paket?", 0, 0, nil)
	if err != nil {
		t.Fatalf("AddMessageWithConfig: %v", err)
	}
	if msg.ID == 0 {
		t.Fatal("expected message to be persisted and get an ID")
	}
	if msg.ConversationID != conv.ID {
		t.Fatalf("expected conversation_id %d, got %d", conv.ID, msg.ConversationID)
	}

	stored := repo.msgs[conv.ID]
	if len(stored) != 1 {
		t.Fatalf("expected 1 stored message, got %d", len(stored))
	}
	if stored[0].Content != "berapa harga paket?" {
		t.Fatalf("unexpected stored content: %q", stored[0].Content)
	}

	// Dengan token + llmConfigID.
	llmID := uint(7)
	botMsg, err := svc.AddMessageWithConfig(ctx, conv.ID, "bot", "Mulai 150rb.", 42, 12, &llmID)
	if err != nil {
		t.Fatalf("AddMessageWithConfig(bot): %v", err)
	}
	if botMsg.TokenIn != 42 || botMsg.TokenOut != 12 {
		t.Fatalf("tokens not persisted: %+v", botMsg)
	}
	if botMsg.LLMConfigID == nil || *botMsg.LLMConfigID != 7 {
		t.Fatalf("llm config id not persisted: %+v", botMsg.LLMConfigID)
	}
}

func TestAddMessageWithConfigValidation(t *testing.T) {
	ctx := context.Background()
	svc := NewConversationUseCase(newFakeRepo())

	if _, err := svc.AddMessageWithConfig(ctx, 0, "customer", "halo", 0, 0, nil); err == nil {
		t.Fatal("expected error for convID == 0")
	}

	repo := newFakeRepo()
	repo.createErr = errors.New("db down")
	failing := NewConversationUseCase(repo)
	if _, err := failing.AddMessageWithConfig(ctx, 1, "customer", "halo", 0, 0, nil); err == nil {
		t.Fatal("expected repo error to propagate")
	}
}

func TestGetRecentHistory(t *testing.T) {
	ctx := context.Background()
	repo := newFakeRepo()
	svc := NewConversationUseCase(repo)

	conv, _ := svc.GetOrCreateConversation(ctx, 1, "628123456789")
	for i := 0; i < 5; i++ {
		if _, err := svc.AddMessageWithConfig(ctx, conv.ID, "customer", "msg", 0, 0, nil); err != nil {
			t.Fatalf("add msg %d: %v", i, err)
		}
	}

	recent, err := svc.GetRecentHistory(ctx, conv.ID, 3)
	if err != nil {
		t.Fatalf("GetRecentHistory: %v", err)
	}
	if len(recent) != 3 {
		t.Fatalf("expected 3 recent messages, got %d", len(recent))
	}
	if recent[0].Content != "msg" || recent[2].Content != "msg" {
		t.Fatal("expected ascending chronological order")
	}
}
