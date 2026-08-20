package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/customer"
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
	seedUsers(ctx, pgStore)
	ensureOwnerExists(ctx, pgStore)
	seedLLMConfig(ctx, pgStore, cfg)
	seedCasbin(ctx, pgStore)

	log.Println("✅ Database seeding completed successfully!")
}

func seedUsers(ctx context.Context, pgStore *postgres.Store) {
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
		existing, err := pgStore.FindUserByUsername(ctx, u.username)
		hash, errHash := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if errHash != nil {
			log.Printf("Failed to hash password for %s: %v", u.username, errHash)
			continue
		}

		if err == nil && existing != nil {
			if err := pgStore.DB().WithContext(ctx).Model(&model.UserModel{}).Where("username = ?", u.username).Update("password_hash", string(hash)).Error; err != nil {
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

		if err := pgStore.CreateUser(ctx, user); err != nil {
			log.Printf("Failed to create user %s: %v", u.username, err)
		} else {
			log.Printf("Created user: %s [Email: %s] (%s)", u.username, u.email, u.role)
		}
	}
}

// ensureOwnerExists menjamin minimal ada satu user ber-role owner — satu-
// satunya role yang bisa mengelola RBAC (rbac:manage). Kalau belum ada
// owner dan ada user admin, admin pertama (ID terkecil) otomatis dinaikkan
// jadi owner. Idempotent: kalau owner sudah ada, tidak melakukan apa-apa.
func ensureOwnerExists(ctx context.Context, pgStore *postgres.Store) {
	users, err := pgStore.FindAllUsers(ctx)
	if err != nil || len(users) == 0 {
		return
	}
	for _, u := range users {
		if u.Role == "owner" {
			return
		}
	}
	for _, u := range users {
		if u.Role == "admin" {
			u.Role = "owner"
			if err := pgStore.UpdateUser(ctx, u); err != nil {
				log.Printf("Failed to promote first admin to owner: %v", err)
			} else {
				log.Printf("Promoted user %q (id=%d) to owner — first admin becomes owner", u.Username, u.ID)
			}
			return
		}
	}
	log.Println("No admin user found to promote to owner — create an owner manually to manage RBAC")
}

func seedCasbin(ctx context.Context, pgStore *postgres.Store) {
	enforcer, err := auth.NewCasbinEnforcer(ctx, pgStore.DB())
	if err != nil {
		log.Printf("Failed to initialize Casbin enforcer for seeding: %v", err)
		return
	}

	auth.SeedSystemPolicies(enforcer)
	log.Println("Seeded full Polyglot system Casbin RBAC policies into Postgres database")

	users, err := pgStore.FindAllUsers(ctx)
	if err != nil {
		log.Printf("Failed to load users for role assignment sync: %v", err)
		return
	}
	refs := make([]*auth.UserRef, 0, len(users))
	for _, u := range users {
		refs = append(refs, &auth.UserRef{ID: fmt.Sprintf("%d", u.ID), Role: u.Role})
	}
	auth.EnsureUserRoleAssignments(enforcer, refs)
	log.Println("Synced user role assignments into Casbin grouping policies")
}

func seedLLMConfig(ctx context.Context, pgStore *postgres.Store, cfg config.Config) {
	configs, _ := pgStore.FindAll(ctx)
	if len(configs) > 0 {
		log.Printf("LLM configs already exist, skipping.")
		return
	}

	encryptedPlaceholder, err := config.Encrypt("REPLACE_WITH_YOUR_GEMINI_API_KEY", cfg.EncryptionKey)
	if err != nil {
		log.Printf("Failed to encrypt default API key: %v", err)
		return
	}

	defaultConfig := &llm.Config{
		Provider:        "gemini",
		Model:           "gemini-2.0-flash",
		APIKeyEncrypted: encryptedPlaceholder,
		MaxOutputTokens: 512,
		IsActive:        true,
	}

	if err := pgStore.Create(ctx, defaultConfig); err != nil {
		log.Printf("Failed to seed LLM config: %v", err)
	} else {
		log.Printf("Created default LLM Config (Google Gemini 2.0 Flash - Active)")
	}
}
