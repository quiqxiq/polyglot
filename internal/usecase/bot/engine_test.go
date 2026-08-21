package bot

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/domain/skill"
	"github.com/quixiq/polyglot/internal/port"
	convUC "github.com/quixiq/polyglot/internal/usecase/conversation"
	skillUC "github.com/quixiq/polyglot/internal/usecase/skill"
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

func TestEngineRateLimitTracking(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "Halo, ada yang bisa dibantu?"}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{ID: 1}}
	convRepo := newFakeConvRepo()

	e := newTestEngine(cache, gw, llmRepo, prov, convRepo)
	ctx := context.Background()
	phone := "628123456789"

	// Awalnya kuota harian = 0
	status, err := e.GetRateLimitStatus(ctx, phone)
	if err != nil {
		t.Fatalf("GetRateLimitStatus error: %v", err)
	}
	if status.DailyChatCount != 0 {
		t.Fatalf("expected initial daily chat count 0, got %d", status.DailyChatCount)
	}

	// Pesan 1 masuk dan dibalas bot
	if err := e.HandleIncomingMessage(ctx, 1, phone+"@s.whatsapp.net", phone, "Halo bot"); err != nil {
		t.Fatalf("HandleIncomingMessage: %v", err)
	}

	// Kuota harian bertambah menjadi 1
	statusAfter, err := e.GetRateLimitStatus(ctx, phone)
	if err != nil {
		t.Fatalf("GetRateLimitStatus error: %v", err)
	}
	if statusAfter.DailyChatCount != 1 {
		t.Fatalf("expected daily chat count 1 after 1 message, got %d", statusAfter.DailyChatCount)
	}

	// Reset rate limit oleh admin
	if err := e.ResetRateLimit(ctx, phone); err != nil {
		t.Fatalf("ResetRateLimit error: %v", err)
	}

	// Kuota harian kembali ke 0
	statusReset, err := e.GetRateLimitStatus(ctx, phone)
	if err != nil {
		t.Fatalf("GetRateLimitStatus error: %v", err)
	}
	if statusReset.DailyChatCount != 0 {
		t.Fatalf("expected daily chat count 0 after reset, got %d", statusReset.DailyChatCount)
	}
}

func TestEngineSkillsInjection(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "Ikuti langkah SOP berikut."}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{
		ID:           1,
		EnableSkills: true,
		SkillsMode:   llm.SkillsModePrompt,
	}}
	convRepo := newFakeConvRepo()

	skillProv := &fakeSkillProvider{
		skills: []skill.SkillInfo{
			{
				Name:        "troubleshoot-los",
				Description: "Panduan lampu LOS merah",
				Content:     "# Langkah Cek Fiber\n1. Pastikan konektor biru kencang.",
			},
		},
	}

	svc := convUC.NewConversationService(convRepo)
	e := NewEngine(
		cache,
		gw,
		svc,
		skillProv,
		&fakeGlobalPromptProvider{prompt: "Kamu adalah asisten GNET."},
		llmRepo,
		newFakeChatRepo(),
		&fakeUserRepo{},
		nil,
		&fakePublisher{},
		func(*llm.Config) (port.LLMProvider, error) { return prov, nil },
	)

	ctx := context.Background()
	phone := "628123456789"
	err := e.HandleIncomingMessage(ctx, 1, phone+"@s.whatsapp.net", phone, "Modem saya lampu merah")
	if err != nil {
		t.Fatalf("HandleIncomingMessage error: %v", err)
	}

	if len(prov.calls) != 1 {
		t.Fatalf("expected 1 LLM call, got %d", len(prov.calls))
	}

	callStr := prov.calls[0]
	assert.Contains(t, callStr, "<available_skills>")
	assert.Contains(t, callStr, "Langkah Cek Fiber")
}

func TestEngineSkillsMode_Tools_LazyLoad(t *testing.T) {
	cache := newFakeCache()
	gw := &fakeGateway{}
	prov := &fakeProvider{reply: "Saya akan mengecek SOP terlebih dahulu."}
	llmRepo := &fakeLLMConfigRepo{active: &llm.Config{
		ID:           1,
		EnableSkills: true,
		SkillsMode:   llm.SkillsModeTools, // Lazy Load Mode!
	}}
	convRepo := newFakeConvRepo()

	skillProv := &fakeSkillProvider{
		skills: []skill.SkillInfo{
			{
				Name:        "ghaib-network-cs",
				Description: "Menangani percakapan CS Ghaib Network",
				Content:     "# SOP Rinci Penanganan Fiber Optik Putus\n1. Ukur OTDR.",
			},
		},
	}

	svc := convUC.NewConversationService(convRepo)
	e := NewEngine(
		cache,
		gw,
		svc,
		skillProv,
		&fakeGlobalPromptProvider{prompt: "Kamu adalah asisten CS."},
		llmRepo,
		newFakeChatRepo(),
		&fakeUserRepo{},
		nil,
		&fakePublisher{},
		func(*llm.Config) (port.LLMProvider, error) { return prov, nil },
	)

	ctx := context.Background()
	phone := "628123456789"
	err := e.HandleIncomingMessage(ctx, 1, phone+"@s.whatsapp.net", phone, "Internet mati total")
	assert.NoError(t, err)

	assert.Len(t, prov.calls, 1)
	callStr := prov.calls[0]

	// 1. Pada mode "tools", full body SOP tidak di-dump di system prompt awal (menghemat token context)
	assert.NotContains(t, callStr, "# SOP Rinci Penanganan Fiber Optik Putus")
	assert.NotContains(t, callStr, "<available_skills>")

	// 2. Tetapi instruksi request_skill diinjeksikan agar LLM dapat memanggil tool on-demand
	assert.Contains(t, callStr, "request_skill")
}

func TestEngineSkills_RequestSkillTool_Execution(t *testing.T) {
	tool := skillUC.RequestSkillTool{
		Skills: []skill.SkillInfo{
			{
				Name:        "ghaib-network-cs",
				Description: "Menangani CS Ghaib Network",
				Content:     "# Panduan Tagihan\nJatuh tempo tanggal 20 setiap bulan.",
			},
		},
	}

	// 1. Sukses mengambil skill yang ada secara on-demand (Lazy Loading)
	result, err := tool.Run(skillUC.RequestSkillArgs{SkillName: "ghaib-network-cs"})
	assert.NoError(t, err)
	assert.Contains(t, result, "Skill 'ghaib-network-cs':")
	assert.Contains(t, result, "# Panduan Tagihan")
	assert.Contains(t, result, "Jatuh tempo tanggal 20 setiap bulan.")

	// 2. Handle skill yang tidak ditemukan
	missingResult, err := tool.Run(skillUC.RequestSkillArgs{SkillName: "unknown-skill"})
	assert.NoError(t, err)
	assert.Contains(t, missingResult, "Skill 'unknown-skill' not found")
	assert.Contains(t, missingResult, "ghaib-network-cs")
}

func TestEngineSkills_SelectedSkillsFilter(t *testing.T) {
	allSkills := []skill.SkillInfo{
		{Name: "skill-a", Description: "Desc A", Content: "Content A"},
		{Name: "skill-b", Description: "Desc B", Content: "Content B"},
		{Name: "skill-c", Description: "Desc C", Content: "Content C"},
	}

	// Jika SelectedSkills kosong, seluruh skill disertakan
	unfiltered := skillUC.FilterSkills(allSkills, nil)
	assert.Len(t, unfiltered, 3)

	// Jika SelectedSkills dispesifikasikan, hanya skill terpilih yang disertakan
	filtered := skillUC.FilterSkills(allSkills, []string{"skill-a", "skill-c"})
	assert.Len(t, filtered, 2)
	assert.Equal(t, "skill-a", filtered[0].Name)
	assert.Equal(t, "skill-c", filtered[1].Name)
}

func TestBotBuiltinTools_GetCurrentTime(t *testing.T) {
	tool := NewGetCurrentTimeTool()
	assert.Equal(t, "get_current_time", tool.Name)
	res, err := tool.Handler(context.Background(), "")
	assert.NoError(t, err)
	assert.Contains(t, res, "Waktu sistem saat ini:")
	assert.Contains(t, res, "WIB")
}

func TestBotBuiltinTools_PingHost(t *testing.T) {
	tool := NewPingHostTool()
	assert.Equal(t, "ping_host", tool.Name)

	// Validasi input kosong
	resEmpty, err := tool.Handler(context.Background(), `{"host":""}`)
	assert.NoError(t, err)
	assert.Contains(t, resEmpty, "Error: parameter 'host'")

	// Ping ke localhost (127.0.0.1)
	res, err := tool.Handler(context.Background(), `{"host":"127.0.0.1", "count":1}`)
	assert.NoError(t, err)
	assert.Contains(t, res, "127.0.0.1")
}

func TestBotBuiltinTools_RequestSkill(t *testing.T) {
	skillProv := &fakeSkillProvider{
		skills: []skill.SkillInfo{
			{
				Name:        "ghaib-network-cs",
				Description: "CS Ghaib Network",
				Content:     "# SOP Tagihan\nJatuh tempo tgl 20.",
			},
		},
	}
	tool := NewRequestSkillTool(skillProv)
	assert.Equal(t, "request_skill", tool.Name)

	// Sukses mengambil skill
	res, err := tool.Handler(context.Background(), `{"skill_name":"ghaib-network-cs"}`)
	assert.NoError(t, err)
	assert.Contains(t, res, "=== SKILL: ghaib-network-cs ===")
	assert.Contains(t, res, "Jatuh tempo tgl 20.")
}

func TestBotBuiltinTools_NotifyTechnician(t *testing.T) {
	gw := &fakeGateway{}
	userRepo := &fakeUserRepo{
		users: []*customer.User{
			{
				ID:             1,
				Username:       "tech_budi",
				FullName:       "Budi Santoso",
				PhoneNumber:    "081249338533",
				Role:           "teknisi",
				Specialization: "Fiber Optic",
				IsActive:       true,
			},
			{
				ID:             2,
				Username:       "tech_agus",
				FullName:       "Agus Pratama",
				PhoneNumber:    "6281234567890",
				Role:           "technician",
				Specialization: "Wireless",
				IsActive:       true,
			},
			{
				ID:             3,
				Username:       "admin_user",
				FullName:       "Admin Sistem",
				PhoneNumber:    "6289999999999",
				Role:           "admin",
				IsActive:       true,
			},
		},
	}
	tool := NewNotifyTechnicianTool(userRepo, gw, 1)
	assert.Equal(t, "notify_technician", tool.Name)

	// 1. Validasi data belum lengkap
	incompleteRes, err := tool.Handler(context.Background(), `{"customer_name":"Budi"}`)
	assert.NoError(t, err)
	assert.Contains(t, incompleteRes, "Error: Data belum lengkap")

	// 2. Data lengkap -> kirim WhatsApp ke teknisi aktif
	completePayload := `{
		"customer_name": "Budi Santoso",
		"customer_phone": "081234567890",
		"address": "Jl. Mawar No. 12 RT 02/03 Sukajadi",
		"issue_type": "Kabel Fiber Putus",
		"issue_description": "Kabel putus tertimpa pohon, modem LOS merah"
	}`
	completeRes, err := tool.Handler(context.Background(), completePayload)
	assert.NoError(t, err)
	assert.Contains(t, completeRes, "Sukses!")
	assert.Contains(t, completeRes, "2 teknisi aktif")

	// Pastikan pesan sampai ke gateway WhatsApp ke-2 teknisi
	assert.Len(t, gw.sent, 2)
	assert.Contains(t, gw.sent[0], "LAPORAN GANGGUAN PELANGGAN")
	assert.Contains(t, gw.sent[0], "Budi Santoso")
	assert.Contains(t, gw.sent[0], "081234567890")
	assert.Contains(t, gw.sent[0], "Jl. Mawar No. 12 RT 02/03 Sukajadi")
	assert.Contains(t, gw.sent[0], "Kabel Fiber Putus")
}

