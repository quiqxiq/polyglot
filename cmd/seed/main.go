package main

import (
	"context"
	"log"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
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
	users := []struct {
		email    string
		password string
		role     string
	}{
		{email: "admin@gnet.co.id", password: "admin123", role: "admin"},
		{email: "agent@gnet.co.id", password: "agent123", role: "agent"},
	}

	for _, u := range users {
		existing, err := pgStore.FindUserByEmail(u.email)
		if err == nil && existing != nil {
			log.Printf("User %s already exists, skipping.", u.email)
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Failed to hash password for %s: %v", u.email, err)
			continue
		}

		user := &customer.User{
			Email:        u.email,
			PasswordHash: string(hash),
			Role:         u.role,
		}

		if err := pgStore.CreateUser(user); err != nil {
			log.Printf("Failed to create user %s: %v", u.email, err)
		} else {
			log.Printf("Created user: %s (%s) [Password: %s]", u.email, u.role, u.password)
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
			Content: "Pembayaran tagihan internet GNET dilakukan paling lambat tanggal 20 setiap bulannya.\n" +
				"Transfer dapat dilakukan ke rekening resmi PT Ghaib Network:\n" +
				"- Bank BCA: 1234-5678-90 a.n PT Ghaib Network\n" +
				"- Bank Mandiri: 137-00-1234567-8 a.n PT Ghaib Network\n" +
				"- Bank BRI: 0012-01-000123-30-5 a.n PT Ghaib Network\n\n" +
				"Setelah melakukan pembayaran, mohon simpan resi transfer dan konfirmasikan ke WhatsApp ini.",
			Tags: "pembayaran,bayar,tagihan,rekening,bca,mandiri,bri,virtual account,va,jatuh tempo,transfer",
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
			Content: "Layanan Pelanggan PT Ghaib Network (GNET) beroperasi 24 jam setiap hari.\n" +
				"- Kantor Pusat: Jl. Jaringan Utama No. 88, Jakarta Selatan\n" +
				"- Telepon: (021) 555-4638\n" +
				"- Email Support: support@gnet.co.id\n" +
				"- Website: https://gnet.co.id",
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
