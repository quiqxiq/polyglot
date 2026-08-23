package config

import (
	_ "embed"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/quixiq/polyglot/pkg/logger"
)

//go:embed rbac_model.conf
var RBACModelConf string

// Config holds foundational infrastructure and environment configuration.
type Config struct {
	Port     string
	AppEnv   string
	LogLevel string

	DatabaseURL string
	RedisURL    string

	EncryptionKey  string
	JWTSecret      string
	JWTExpiryHours int

	RefreshTokenTTLHours int

	WAStoreDir string

	CORSOrigins []string

	// Scheduler ISP (fase 3).
	BillingCronSpec   string
	IsolationCronSpec string
	SchedulerEnabled  bool
}

// Load reads infrastructure secrets and settings from environment variables or .env files.
func Load() Config {
	if err := godotenv.Load(); err != nil {
		logger.WithComponent("Config").Warn("No .env file found or error loading, using environment/default values")
	}
	_ = godotenv.Overload(".env.local")

	originsRaw := getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173,http://127.0.0.1:3000")
	var origins []string
	if originsRaw == "*" {
		origins = []string{"*"}
	} else {
		for _, o := range strings.Split(originsRaw, ",") {
			trimmed := strings.TrimSpace(o)
			if trimmed != "" {
				origins = append(origins, trimmed)
			}
		}
	}

	return Config{
		Port:                 getEnv("PORT", "8080"),
		AppEnv:               getEnv("APP_ENV", "development"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		DatabaseURL:          getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/polyglot?sslmode=disable"),
		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379/0"),
		EncryptionKey:        getEnv("ENCRYPTION_KEY", "12345678901234567890123456789012"),
		JWTSecret:            getEnv("JWT_SECRET", "polyglot-super-secret-jwt-key-change-in-production"),
		JWTExpiryHours:       getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		RefreshTokenTTLHours: getEnvAsInt("REFRESH_TOKEN_TTL_HOURS", 168),
		WAStoreDir:           getEnv("WA_STORE_DIR", "./data/whatsapp"),
		CORSOrigins:          origins,

		// Scheduler ISP (fase 3). Kosongkan spesifikasi untuk menonaktifkan
		// job terkait.
		BillingCronSpec:   getEnv("BILLING_CRON", "0 6 * * *"),
		IsolationCronSpec: getEnv("ISOLATION_CRON", "@every 10m"),
		SchedulerEnabled:  getEnv("SCHEDULER_ENABLED", "true") == "true",
	}
}

func (c *Config) Validate() error {
	if len(c.EncryptionKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes (got %d)", len(c.EncryptionKey))
	}
	if len(c.JWTSecret) < 16 {
		return fmt.Errorf("JWT_SECRET must be at least 16 characters for security")
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	valStr := getEnv(key, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}
