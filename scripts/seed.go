package main

import (
	"context"
	"log"

	"golang.org/x/crypto/bcrypt"

	"github.com/quixiq/polyglot/internal/adapter/auth"
	"github.com/quixiq/polyglot/internal/adapter/postgres"
	"github.com/quixiq/polyglot/internal/adapter/postgres/models"
	"github.com/quixiq/polyglot/internal/config"
	"github.com/quixiq/polyglot/internal/domain/customer"
)

func main() {
	cfg := config.Load()

	log.Printf("[Seeder] Connecting to PostgreSQL database...")
	pgStore, err := postgres.NewStore(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[Seeder Error] Failed to connect to DB: %v", err)
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
			log.Fatalf("[Seeder Error] Failed to hash password: %v", err)
		}

		adminUser := &customer.User{
			Email:        adminEmail,
			PasswordHash: string(hash),
			Role:         "admin",
			TenantID:     "tenant-default",
		}

		if err := pgStore.CreateUser(adminUser); err != nil {
			log.Fatalf("[Seeder Error] Failed to create admin user: %v", err)
		}
		log.Printf("[Seeder] Created default Admin user: %s (Password: %s)", adminEmail, adminPassword)
	} else {
		log.Printf("[Seeder] Admin user %s already exists in database.", adminEmail)
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
