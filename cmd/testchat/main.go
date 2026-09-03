package main

import (
	"context"
	"fmt"
	"time"

	"github.com/quixiq/polyglot/internal/adapter/llm"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/redis"
	"github.com/quixiq/polyglot/internal/adapter/storage"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/bot"
	"github.com/quixiq/polyglot/internal/domain/customer"
	domainLLM "github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/internal/port"
	usecaseBot "github.com/quixiq/polyglot/internal/usecase/bot"
	usecaseConv "github.com/quixiq/polyglot/internal/usecase/conversation"
	usecaseSkill "github.com/quixiq/polyglot/internal/usecase/skill"
	"github.com/quixiq/polyglot/pkg/logger"
)

type mockGateway struct {
	lastReply    string
	sentMessages map[string][]string
}

func (m *mockGateway) Connect(*bot.WASession) error { return nil }
func (m *mockGateway) Disconnect(uint) error        { return nil }
func (m *mockGateway) Logout(uint) error            { return nil }
func (m *mockGateway) Purge(uint) error             { return nil }
func (m *mockGateway) Reconnect(uint) error         { return nil }
func (m *mockGateway) SendMessage(sessionID uint, to string, content string) error {
	m.lastReply = content
	if m.sentMessages == nil {
		m.sentMessages = make(map[string][]string)
	}
	m.sentMessages[to] = append(m.sentMessages[to], content)
	return nil
}
func (m *mockGateway) SendDocument(context.Context, uint, string, []byte, string, string, string) error {
	return nil
}
func (m *mockGateway) SendImage(context.Context, uint, string, []byte, string, string) error {
	return nil
}
func (m *mockGateway) GetStatus(uint) (string, error)              { return "online", nil }
func (m *mockGateway) GetQRCode(uint) (string, error)              { return "", nil }
func (m *mockGateway) GetPairingCode(uint, string) (string, error) { return "", nil }
func (m *mockGateway) RestoreAllSessions([]bot.WASession) error    { return nil }

type mockPublisher struct{}

func (p *mockPublisher) PublishEvent(eventType string, payload any) {}

func main() {
	logger.Init("info", "development")
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		logger.WithComponent("TestChat").WithError(err).Fatal("invalid configuration")
	}

	ctx := context.Background()
	pgStore, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		logger.WithComponent("TestChat").WithError(err).Fatal("failed to connect to postgres")
	}

	redisStore, err := redis.NewStore(cfg.RedisURL)
	if err != nil {
		logger.WithComponent("TestChat").WithError(err).Fatal("failed to connect to redis")
	}

	fsSkillStore, err := storage.NewFSSkillStore("data")
	if err != nil {
		logger.WithComponent("TestChat").WithError(err).Fatal("failed to init skill store")
	}

	skillUseCase := usecaseSkill.NewManageSkillUseCase(pgStore, fsSkillStore, nil)
	convService := usecaseConv.NewConversationUseCase(pgStore)

	gw := &mockGateway{
		sentMessages: make(map[string][]string),
	}
	publisher := &mockPublisher{}

	providerFactory := func(llmCfg *domainLLM.Config) (port.LLMProvider, error) {
		return llm.NewProvider(llmCfg, cfg.EncryptionKey)
	}

	userRepo := postgres.NewUserRepository(pgStore.DB())
	settingRepo := postgres.NewSettingRepository(pgStore.DB())
	engine := usecaseBot.NewEngine(
		redisStore,
		gw,
		convService,
		skillUseCase.Provider(),
		skillUseCase,
		pgStore,
		pgStore,
		userRepo,
		settingRepo,
		publisher,
		providerFactory,
	)

	// Ensure WhatsApp test session exists in DB
	sess, err := pgStore.FindSessionByID(ctx, 1)
	if err != nil || sess == nil {
		testSess := &bot.WASession{
			ID:           1,
			DeviceName:   "Live Test Session",
			PhoneNumber:  "628111222333",
			Status:       bot.StatusOnline,
			IsBotEnabled: true,
		}
		_ = pgStore.CreateSession(ctx, testSess)
	}

	// Ensure active technician user exists in DB
	techs, _ := pgStore.FindUsersByRoles(ctx, []string{"teknisi", "technician"}, true)
	if len(techs) == 0 {
		testTech := &customer.User{
			Username:       "tech_lapangan",
			Email:          "tech@gnet.local",
			PasswordHash:   "$2a$10$testHashPlaceholder",
			Role:           "teknisi",
			FullName:       "Budi Santoso (Teknisi Lapangan)",
			PhoneNumber:    "6281249338533",
			Specialization: "Fiber Optic & Splicing",
			IsActive:       true,
			TenantID:       "tenant-default",
		}
		_ = pgStore.CreateUser(ctx, testTech)
	}

	// Baca active LLM config dari database
	activeConfig, err := pgStore.FindActive(ctx)
	if err != nil {
		logger.WithComponent("TestChat").WithError(err).Fatal("no active llm config found")
	}
	fmt.Printf("=== [LIVE TEST] Active LLM Config: ID=%d, Provider=%s, Model=%s, EnableSkills=%v, SkillsMode=%s ===\n",
		activeConfig.ID, activeConfig.Provider, activeConfig.Model, activeConfig.EnableSkills, activeConfig.SkillsMode)

	// Pastikan skill aktif dan model tervalidasi
	activeConfig.EnableSkills = true
	activeConfig.SkillsMode = "prompt"
	activeConfig.Model = "openai/gpt-oss-120b"
	_ = pgStore.Update(ctx, activeConfig)

	// Skenario 1: Pertanyaan gangguan koneksi (SOP troubleshooting-jaringan.md)
	fmt.Println("\n-------------------------------------------------------------")
	fmt.Println("[TEST SKENARIO 1] Pelanggan: 'Halo kak, internet saya mati total dan lampu modem saya warna merah, bagaimana solusinya?'")
	fmt.Println("-------------------------------------------------------------")

	customerPhone := fmt.Sprintf("628%d", time.Now().Unix())
	chatJID := customerPhone + "@s.whatsapp.net"

	err = engine.HandleIncomingMessage(ctx, 1, chatJID, customerPhone, "Halo kak, internet saya mati total dan lampu modem saya warna merah, bagaimana solusinya?")
	if err != nil {
		logger.WithComponent("TestChat").WithError(err).Fatal("handle incoming message failed")
	}

	fmt.Println("\n[HASIL JAWABAN BOT (Menggunakan SOP Ghaib Network)]:")
	fmt.Println(gw.lastReply)

	// Skenario 2: Pertanyaan seputar hal di luar cakupan (SOP profil-perusahaan.md / batasan)
	fmt.Println("\n-------------------------------------------------------------")
	fmt.Println("[TEST SKENARIO 2] Pelanggan: 'Kak, bisa tolong buatkan website toko online dan pasang CCTV di rumah saya?'")
	fmt.Println("-------------------------------------------------------------")

	err = engine.HandleIncomingMessage(ctx, 1, chatJID, customerPhone, "Kak, bisa tolong buatkan website toko online dan pasang CCTV di rumah saya?")
	if err != nil {
		logger.WithComponent("TestChat").WithError(err).Fatal("handle incoming message scenario 2 failed")
	}

	fmt.Println("\n[HASIL JAWABAN BOT (Pengalihan di Luar Cakupan SOP)]:")
	fmt.Println(gw.lastReply)

	// Skenario 3: Pelaporan gangguan fisik & permintaan kunjungan teknisi
	fmt.Println("\n-------------------------------------------------------------")
	fmt.Println("[TEST SKENARIO 3] Pelanggan Melaporkan Data Lengkap untuk Kunjungan Teknisi:")
	fmt.Println("'Kabel optik rumah saya putus tertimpa dahan pohon kak. Tolong kirim teknisi ke rumah. Nama saya Budi Santoso, alamat di Jl. Mawar No. 12 RT 02/03 Sukajadi, no HP aktif 081234567890'")
	fmt.Println("-------------------------------------------------------------")

	techCustomerPhone := fmt.Sprintf("628%d", time.Now().Unix()+10)
	techChatJID := techCustomerPhone + "@s.whatsapp.net"

	err = engine.HandleIncomingMessage(ctx, 1, techChatJID, techCustomerPhone,
		"Kabel optik rumah saya putus tertimpa dahan pohon kak. Tolong kirim teknisi ke rumah. Nama saya Budi Santoso, alamat di Jl. Mawar No. 12 RT 02/03 Sukajadi, no HP aktif 081234567890",
	)
	if err != nil {
		logger.WithComponent("TestChat").WithError(err).Fatal("handle incoming message scenario 3 failed")
	}

	fmt.Println("\n[HASIL JAWABAN BOT KE PELANGGAN (Turn 1)]:")
	fmt.Println(gw.lastReply)

	// Turn 2: Pelanggan melengkapi konfirmasi
	fmt.Println("\n[TEST SKENARIO 3 - Turn 2] Pelanggan: 'Kelurahan Sukajadi, Kecamatan Sukajadi, Kota Bandung kak. Tolong segera diteruskan ke teknisi ya.'")
	err = engine.HandleIncomingMessage(ctx, 1, techChatJID, techCustomerPhone,
		"Kelurahan Sukajadi, Kecamatan Sukajadi, Kota Bandung kak. Tolong segera diteruskan ke teknisi ya.",
	)
	if err != nil {
		logger.WithComponent("TestChat").WithError(err).Fatal("handle incoming message turn 2 failed")
	}

	fmt.Println("\n[HASIL JAWABAN BOT KE PELANGGAN (Turn 2)]:")
	fmt.Println(gw.lastReply)

	// Cek apakah pesan ke WhatsApp teknisi terkirim
	fmt.Println("\n[PESAN YANG DITERIMA WHATSAPP TEKNISI / GATEWAY]:")
	for jid, msgs := range gw.sentMessages {
		if len(msgs) > 0 {
			fmt.Printf("-> Target %s (%d messages):\n", jid, len(msgs))
			fmt.Println(msgs[len(msgs)-1])
		}
	}

	fmt.Println("\n=== LIVE CHATBOT TEST COMPLETED SUCCESSFULLY! ===")
}
