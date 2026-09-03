package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/model"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/customer"
	"github.com/quixiq/polyglot/internal/domain/llm"
	"github.com/quixiq/polyglot/pkg/logger"
)

func main() {
	if err := godotenv.Load(); err != nil {
		logger.WithComponent("Seeder").Debug("no .env file found, using default environment variables")
	}

	cfg := config.Load()
	logger.Init(cfg.LogLevel, cfg.AppEnv)
	if err := cfg.Validate(); err != nil {
		logger.WithComponent("Seeder").WithError(err).Fatal("invalid configuration")
	}

	logger.WithComponent("Seeder").Info("connecting to database for seeding")

	pgStore, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		logger.WithComponent("Seeder").WithError(err).Fatal("database connection failed")
	}

	logger.WithComponent("Seeder").Info("seeding database data")

	ctx := context.Background()
	seedUsers(ctx, pgStore)
	seedLLMConfig(ctx, pgStore, cfg)
	seedCasbin(ctx, pgStore)

	logger.WithComponent("Seeder").Info("database seeding completed successfully")
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
			logger.WithComponent("Seeder").WithField("username", u.username).Fatal("password does not meet strength requirements (min 8 chars)")
		}

		existing, err := pgStore.FindUserByUsername(ctx, u.username)
		hash, errHash := bcrypt.GenerateFromPassword([]byte(u.password), bcrypt.DefaultCost)
		if errHash != nil {
			logger.WithComponent("Seeder").WithError(errHash).WithField("username", u.username).Error("failed to hash password")
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
				logger.WithComponent("Seeder").WithError(err).WithField("username", u.username).Error("failed to update existing user")
			} else {
				logger.WithComponent("Seeder").WithFields(map[string]any{
					"username": u.username,
					"role":     u.role,
					"email":    u.email,
				}).Info("updated existing user")
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
			TenantID:       "tenant-default",
		}

		if err := pgStore.CreateUser(ctx, user); err != nil {
			logger.WithComponent("Seeder").WithError(err).WithField("username", u.username).Error("failed to create user")
		} else {
			logger.WithComponent("Seeder").WithFields(map[string]any{
				"username": u.username,
				"role":     u.role,
				"email":    u.email,
			}).Info("created user")
		}
	}
}

func seedCasbin(ctx context.Context, pgStore *postgres.Store) {
	enforcer, err := auth.NewCasbinEnforcer(ctx, pgStore.DB())
	if err != nil {
		logger.WithComponent("Seeder").WithError(err).Error("failed to initialize casbin enforcer for seeding")
		return
	}

	auth.SeedSystemPolicies(enforcer)
	logger.WithComponent("Seeder").Info("seeded system casbin RBAC policies into postgres database")

	users, err := pgStore.FindAllUsers(ctx)
	if err != nil {
		logger.WithComponent("Seeder").WithError(err).Error("failed to load users for role assignment sync")
		return
	}
	refs := make([]*auth.UserRef, 0, len(users))
	for _, u := range users {
		refs = append(refs, &auth.UserRef{ID: fmt.Sprintf("%d", u.ID), Role: u.Role})
	}
	auth.EnsureUserRoleAssignments(enforcer, refs)
	logger.WithComponent("Seeder").Info("synced user role assignments into casbin grouping policies")
}

func seedLLMConfig(ctx context.Context, pgStore *postgres.Store, cfg config.Config) {
	configs, _ := pgStore.FindAll(ctx)
	if len(configs) > 0 {
		logger.WithComponent("Seeder").Info("llm configs already exist, skipping")
		return
	}

	encryptedPlaceholder, err := config.Encrypt("REPLACE_WITH_YOUR_GEMINI_API_KEY", cfg.EncryptionKey)
	if err != nil {
		logger.WithComponent("Seeder").WithError(err).Error("failed to encrypt default api key")
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
		logger.WithComponent("Seeder").WithError(err).Error("failed to seed llm config")
	} else {
		logger.WithComponent("Seeder").WithFields(map[string]any{
			"provider": defaultConfig.Provider,
			"model":    defaultConfig.Model,
		}).Info("created default llm config")
	}
}
