package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/knowledge"
	"github.com/quixiq/polyglot/internal/domain/llm"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using default environment variables")
	}

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}

	log.Println("Connecting to database for seeding...")

	pgStore, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	log.Println("Seeding database data...")

	ctx := context.Background()
	seedUsers(pgStore)
	seedKnowledge(pgStore)
	seedLLMConfig(pgStore, cfg)
	seedCasbin(ctx, pgStore)

	log.Println("✅ Database seeding completed successfully!")
}

func seedUsers(pgStore *postgres.Store) {
	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Println("SEED_ADMIN_PASSWORD not set, skipping admin user seeding.")
	}
	agentPassword := os.Getenv("SEED_AGENT_PASSWORD")
	if agentPassword == "" {
		log.Println("SEED_AGENT_PASSWORD not set, skipping agent user seeding.")
	}

	if adminPassword == "" && agentPassword == "" {
		return
	}

	validatePassword := func(pw string) bool {
		return len(pw) >= 8
	}

	users := []struct {
		username string
		email    string
		password string
		role     string
	}{}

	if adminPassword != "" {
		if !validatePassword(adminPassword) {
			log.Fatalf("SEED_ADMIN_PASSWORD does not meet strength requirements (min 8 chars)")
		}
		users = append(users, struct {
			username string
			email    string
			password string
			role     string
		}{username: "admin", email: "admin@example.com", password: adminPassword, role: "admin"})
	}

	if agentPassword != "" {
		if !validatePassword(agentPassword) {
			log.Fatalf("SEED_AGENT_PASSWORD does not meet strength requirements (min 8 chars)")
		}
		users = append(users, struct {
			username string
			email    string
			password string
			role     string
		}{username: "agent", email: "agent@example.com", password: agentPassword, role: "agent"})
	}

	for _, u := range users {
		existing, err := pgStore.FindUserByUsername(u.username)
		hash, errHash := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if errHash != nil {
			log.Printf("Failed to hash password for %s: %v", u.username, errHash)
			continue
		}

		if err == nil && existing != nil {
			if err := pgStore.DB().Model(&models.UserModel{}).Where("username = ?", u.username).Update("password_hash", string(hash)).Error; err != nil {
				log.Printf("Failed to update password for existing user %s: %v", u.username, err)
			} else {
				log.Printf("Updated password for existing user: %s", u.username)
			}
			continue
		}

		user := &customer.User{
			Username:     u.username,
			Email:        u.email,
			PasswordHash: string(hash),
			Role:         u.role,
		}

		if err := pgStore.CreateUser(user); err != nil {
			log.Printf("Failed to create user %s: %v", u.username, err)
		} else {
			log.Printf("Created user: %s [Email: %s] (%s)", u.username, u.email, u.role)
		}
	}
}

func seedCasbin(ctx context.Context, pgStore *postgres.Store) {
	enforcer, err := auth.NewCasbinEnforcer(ctx, pgStore.DB())
	if err != nil {
		log.Printf("Failed to initialize Casbin enforcer for seeding: %v", err)
		return
	}

	auth.SeedSystemPolicies(enforcer)
	log.Println("Seeded full Polyglot system Casbin RBAC policies into Postgres database")
}

func seedKnowledge(pgStore *postgres.Store) {
	entries := []knowledge.KnowledgeEntry{
		{
			Title: "Paket & Harga Internet GNET Home",
			Content: "PT Ghaib Network (GNET) menyediakan paket internet rumah unmetered unlimited tanpa FUP:\n" +
				"1. GNET Starter 20 Mbps — Rp 199.000 / bulan\n" +
				"2. GNET Family 50 Mbps — Rp 299.000 / bulan\n" +
				"3. GNET Ultra 100 Mbps — Rp 499.000 / bulan\n\n" +
				"Harga sudah termasuk PPN 11% dan sewa router modem Wi-Fi dual-band gratis.",
			Tags: "paket,harga,tarif,biaya,kecepatan,mbps,home,internet,unlimited,fup,promo",
		},
		{
			Title: "Prosedur Pembayaran Tagihan Bulanan",
			Content: "Pembayaran tagihan internet dilakukan paling lambat tanggal 20 setiap bulannya.\n" +
				"Transfer dapat dilakukan ke rekening resmi perusahaan yang tertera pada invoice.\n" +
				"Setelah melakukan pembayaran, mohon simpan resi transfer dan konfirmasikan ke WhatsApp ini.",
			Tags: "pembayaran,bayar,tagihan,rekening,virtual account,va,jatuh tempo,transfer",
		},
		{
			Title: "Penanganan Gangguan Koneksi Internet (Troubleshooting)",
			Content: "Jika jaringan internet GNET Anda mengalami gangguan atau lampu indikator LOS berwarna merah:\n" +
				"1. Matikan router/modem GNET selama 30 detik lalu nyalakan kembali.\n" +
				"2. Pastikan kabel fiber optik berwarna kuning terpasang kencang dan tidak tertekuk.\n" +
				"3. Jika kendala masih berlanjut, infokan lokasi dan nama pelanggan. Tim teknisi GNET siap datang ke lokasi.",
			Tags: "gangguan,rusak,los,merah,lambat,lemot,mati,trouble,restart,modem,router,teknisi,kabel",
		},
		{
			Title: "Pemasangan Baru & Cek Coverage Area",
			Content: "Untuk mendaftar pemasangan baru internet GNET:\n" +
				"1. Kirimkan lokasi Share Location (Google Maps) rumah/kantor Anda untuk cek ODP terdekat.\n" +
				"2. Pemasangan gratis biaya instalasi (Free Biaya Pasang).\n" +
				"3. Proses instalasi dilakukan oleh teknisi dalam waktu 1x24 jam setelah verifikasi.",
			Tags: "pemasangan,pasang,baru,coverage,area,jangkauan,odp,lokasi,registrasi,daftar,instalasi",
		},
		{
			Title: "Kontak Layanan Pelanggan & Kantor Pusat",
			Content: "Layanan pelanggan beroperasi 24 jam setiap hari.\n" +
				"- Telepon: sesuai nomor yang tertera pada invoice atau aplikasi\n" +
				"- Email Support: support@example.com\n" +
				"- Website: https://example.com",
			Tags: "kontak,alamat,lokasi,kantor,telepon,email,call center,operasional,jam,hubungi",
		},
	}

	existingEntries, _ := pgStore.FindAllKnowledgeEntries()
	if len(existingEntries) > 0 {
		log.Printf("Knowledge base already has %d entries, skipping.", len(existingEntries))
		return
	}

	for i := range entries {
		entry := &entries[i]
		if err := pgStore.CreateKnowledgeEntry(entry); err != nil {
			log.Printf("Failed to seed knowledge entry '%s': %v", entry.Title, err)
		} else {
			log.Printf("Created knowledge entry: %s", entry.Title)
		}
	}
}

func seedLLMConfig(pgStore *postgres.Store, cfg config.Config) {
	configs, _ := pgStore.FindAllLLMConfigs()
	if len(configs) > 0 {
		log.Printf("LLM configs already exist, skipping.")
		return
	}

	encryptedPlaceholder, err := config.Encrypt("REPLACE_WITH_YOUR_GEMINI_API_KEY", cfg.EncryptionKey)
	if err != nil {
		log.Printf("Failed to encrypt default API key: %v", err)
		return
	}

	defaultConfig := &llm.LLMConfig{
		Provider:        "gemini",
		Model:           "gemini-2.0-flash",
		APIKeyEncrypted: encryptedPlaceholder,
		MaxOutputTokens: 512,
		IsActive:        true,
	}

	if err := pgStore.CreateLLMConfig(defaultConfig); err != nil {
		log.Printf("Failed to seed LLM config: %v", err)
	} else {
		log.Printf("Created default LLM Config (Google Gemini 2.0 Flash - Active)")
	}
}
