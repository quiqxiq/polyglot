package main

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/config"
	domainAuth "github.com/quixiq/polyglot/internal/domain/auth"
	"github.com/quixiq/polyglot/pkg/logger"
)

func main() {
	cfg := config.Load()

	logger.Info("[Seeder] Connecting to PostgreSQL database...")
	pgStore, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		logger.WithError(err).Fatal("[Seeder Error] Failed to connect to DB")
	}

	db := pgStore.DB()

	// 1. Seed Admin User
	adminEmail := "admin@gnet.id"
	adminPassword := "admin123"

	var existing models.UserModel
	err = db.Where("email = ?", adminEmail).First(&existing).Error
	if err != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			logger.WithError(err).Fatal("[Seeder Error] Failed to hash password")
		}

		adminUser := &domainAuth.User{
			Email:        adminEmail,
			PasswordHash: string(hash),
			Role:         "admin",
			TenantID:     "tenant-default",
		}

		if err := pgStore.CreateUser(adminUser); err != nil {
			logger.WithError(err).Fatal("[Seeder Error] Failed to create admin user")
		}
		logger.Infof("[Seeder] Created default Admin user: %s (Password: %s)", adminEmail, adminPassword)
	} else {
		logger.Infof("[Seeder] Admin user %s already exists in database.", adminEmail)
	}

	// 2. Initialize Casbin Enforcer & Seed System RBAC Policies
	ctx := context.Background()
	enforcer, err := auth.NewCasbinEnforcer(ctx, db)
	if err != nil {
		logger.WithError(err).Warn("[Seeder Warning] Failed to initialize Casbin enforcer")
	} else {
		auth.SeedSystemPolicies(enforcer)
		_, _ = enforcer.AddRoleForUser(adminEmail, "admin")
		logger.Infof("[Seeder] Assigned role 'admin' to user %s in Casbin", adminEmail)
	}

	logger.Info("[Seeder] Database seeding completed successfully!")
}
