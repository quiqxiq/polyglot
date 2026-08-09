package main

import (
	"context"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

func main() {
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("[Seeder Error] invalid configuration: %v", err)
	}

	log.Printf("[Seeder] Connecting to PostgreSQL database...")
	pgStore, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[Seeder Error] Failed to connect to DB: %v", err)
	}

	db := pgStore.DB()

	// 1. Seed Admin User
	adminUsername := "admin"
	adminEmail := "admin@example.com"
	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	if adminPassword == "" {
		log.Fatalf("[Seeder Error] SEED_ADMIN_PASSWORD environment variable is required")
	}
	if len(adminPassword) < 8 {
		log.Fatalf("[Seeder Error] SEED_ADMIN_PASSWORD must be at least 8 characters long")
	}

	var existing models.UserModel
	err = db.Where("username = ?", adminUsername).First(&existing).Error
	if err != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("[Seeder Error] Failed to hash password: %v", err)
		}

		adminUser := &customer.User{
			Username:     adminUsername,
			Email:        adminEmail,
			PasswordHash: string(hash),
			Role:         "admin",
			TenantID:     "tenant-default",
		}

		if err := pgStore.CreateUser(adminUser); err != nil {
			log.Fatalf("[Seeder Error] Failed to create admin user: %v", err)
		}
		log.Printf("[Seeder] Created default Admin user: %s", adminUsername)
	} else {
		log.Printf("[Seeder] Admin user %s already exists in database.", adminUsername)
	}

	// 2. Initialize Casbin Enforcer & Seed System RBAC Policies
	ctx := context.Background()
	enforcer, err := auth.NewCasbinEnforcer(ctx, db)
	if err != nil {
		log.Printf("[Seeder Warning] Failed to initialize Casbin enforcer: %v", err)
	} else {
		auth.SeedSystemPolicies(enforcer)
		_, _ = enforcer.AddRoleForUser(adminEmail, "admin")
		log.Printf("[Seeder] Assigned role 'admin' to user %s in Casbin", adminEmail)
	}

	log.Println("[Seeder] Database seeding completed successfully!")
}
