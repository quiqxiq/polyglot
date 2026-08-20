package bot

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/llm"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
)

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
	if err := svc.TakeOver(context.Background(), conv.ID, 1); err != nil {
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
		if c.Status != bot.StatusBot {
			t.Fatalf("expected conversation status bot, got %s", c.Status)
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

	// LLM gagal → fallback ramah dikirim dan status tetap 'bot' agar pesan berikutnya tidak terkunci.
	if len(gw.sent) != 1 {
		t.Fatalf("expected fallback reply, got %v", gw.sent)
	}
	var convID uint
	for id, c := range convRepo.convs {
		convID = id
		if c.Status != bot.StatusBot {
			t.Fatalf("expected conversation to remain in bot status after transient LLM error, got %s", c.Status)
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

func TestEngineIgnoresGroupsAndBroadcasts(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "balasan"}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()

	e := newTestEngine(cache, gw, llmRepo, prov, convRepo)

	// 1. Group message
	if err := e.HandleIncomingMessage(context.Background(), 1, "123456789-987654@g.us", "628123456789", "Halo semua di grup"); err != nil {
		t.Fatalf("unexpected error for group: %v", err)
	}

	// 2. Broadcast / status
	if err := e.HandleIncomingMessage(context.Background(), 1, "status@broadcast", "628123456789", "Status update"); err != nil {
		t.Fatalf("unexpected error for broadcast: %v", err)
	}

	// 3. Channel / Newsletter
	if err := e.HandleIncomingMessage(context.Background(), 1, "123456789012345678@newsletter", "news", "Berita terbaru"); err != nil {
		t.Fatalf("unexpected error for newsletter: %v", err)
	}

	// Pastikan bot tidak pernah mengirim balasan dan tidak memanggil LLM
	if len(gw.sent) != 0 {
		t.Fatalf("bot must not reply to groups or broadcasts, got %v", gw.sent)
	}
	if len(prov.calls) != 0 {
		t.Fatalf("LLM must not be called for groups or broadcasts")
	}
}
