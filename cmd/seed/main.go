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
	seedLLMConfig(ctx, pgStore, cfg)
	seedCasbin(ctx, pgStore)

	log.Println("✅ Database seeding completed successfully!")
}

func seedUsers(ctx context.Context, pgStore *postgres.Store) {
	ownerPassword := os.Getenv("SEED_OWNER_PASSWORD")
	if ownerPassword == "" {
		ownerPassword = "owner12345"
	}
	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "admin12345"
	}
	agentPassword := os.Getenv("SEED_AGENT_PASSWORD")
	if agentPassword == "" {
		agentPassword = "agent123"
	}

	validatePassword := func(pw string) bool {
		return len(pw) >= 8
	}

	users := []struct {
		username       string
		email          string
		password       string
		role           string
		fullName       string
		phoneNumber    string
		specialization string
	}{
		{
			username:       "owner",
			email:          "owner@example.com",
			password:       ownerPassword,
			role:           "owner",
			fullName:       "System Owner",
			phoneNumber:    "628111111111",
			specialization: "Superadmin",
		},
		{
			username:       "admin",
			email:          "admin@example.com",
			password:       adminPassword,
			role:           "admin",
			fullName:       "Administrator",
			phoneNumber:    "628222222222",
			specialization: "Network Admin",
		},
		{
			username:       "agent",
			email:          "agent@example.com",
			password:       agentPassword,
			role:           "agent",
			fullName:       "Customer Service",
			phoneNumber:    "628333333333",
			specialization: "Billing & CS",
		},
		{
			username:       "tech_lapangan",
			email:          "tech@gnet.local",
			password:       "teknisi123",
			role:           "teknisi",
			fullName:       "Budi Santoso",
			phoneNumber:    "6281249338533",
			specialization: "Fiber Optic",
		},
	}

	for _, u := range users {
		if !validatePassword(u.password) {
			log.Fatalf("Password for %s does not meet strength requirements (min 8 chars)", u.username)
		}

		existing, err := pgStore.FindUserByUsername(ctx, u.username)
		hash, errHash := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if errHash != nil {
			log.Printf("Failed to hash password for %s: %v", u.username, errHash)
			continue
		}

		if err == nil && existing != nil {
			updates := map[string]any{
				"password_hash":  string(hash),
				"role":           u.role,
				"full_name":      u.fullName,
				"phone_number":   u.phoneNumber,
				"specialization": u.specialization,
				"is_active":      true,
			}
			if err := pgStore.DB().WithContext(ctx).Model(&model.UserModel{}).Where("username = ?", u.username).Updates(updates).Error; err != nil {
				log.Printf("Failed to update existing user %s: %v", u.username, err)
			} else {
				log.Printf("Updated existing user: %s [Role: %s, Email: %s]", u.username, u.role, u.email)
			}
			continue
		}

		user := &customer.User{
			Username:       u.username,
			Email:          u.email,
			PasswordHash:   string(hash),
			Role:           u.role,
			FullName:       u.fullName,
			PhoneNumber:    u.phoneNumber,
			Specialization: u.specialization,
			IsActive:       true,
		}

		if err := pgStore.CreateUser(ctx, user); err != nil {
			log.Printf("Failed to create user %s: %v", u.username, err)
		} else {
			log.Printf("Created user: %s [Role: %s, Email: %s]", u.username, u.role, u.email)
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
