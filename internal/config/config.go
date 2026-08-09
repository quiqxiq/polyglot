package config

import (
	_ "embed"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

//go:embed rbac_model.conf
var RBACModelConf string

//go:embed system_prompt.md
var DefaultSystemPrompt string

type Config struct {
	Port   string
	AppEnv string

	DatabaseURL string
	RedisURL    string

	EncryptionKey  string
	JWTSecret      string
	JWTExpiryHours int

	SessionTimeoutMinutes int
	SlidingWindowSize     int

	RateLimitPerMinute int
	RateLimitPerHour   int

	LLMMaxOutputTokens int

	AllowedTopics []string
	SystemPrompt  string

	WAStoreDir string

	CORSOrigins []string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found or error loading, using environment/default values")
	}
	// Per-developer overrides loaded after .env so they take precedence.
	_ = godotenv.Load(".env.local")

	return Config{
		Port:   getEnv("APP_PORT", getEnv("PORT", "8080")),
		AppEnv: getEnv("APP_ENV", "development"),

		DatabaseURL: getEnv("DATABASE_URL", ""),
		RedisURL:    getEnv("REDIS_URL", "redis://localhost:6379/0"),

		EncryptionKey:  getEnv("ENCRYPTION_KEY", ""),
		JWTSecret:      getEnv("JWT_SECRET", ""),
		JWTExpiryHours: getEnvInt("JWT_EXPIRY_HOURS", 24),

		SessionTimeoutMinutes: getEnvInt("SESSION_TIMEOUT_MINUTES", 30),
		SlidingWindowSize:     getEnvInt("SLIDING_WINDOW_SIZE", 10),

		RateLimitPerMinute: getEnvInt("RATE_LIMIT_PER_MINUTE", 10),
		RateLimitPerHour:   getEnvInt("RATE_LIMIT_PER_HOUR", 60),

		LLMMaxOutputTokens: getEnvInt("LLM_MAX_OUTPUT_TOKENS", 512),

		AllowedTopics: getEnvSlice("ALLOWED_TOPICS", []string{
			"layanan internet", "paket", "harga", "tagihan",
			"gangguan", "pemasangan", "coverage", "area",
			"pembayaran", "promo", "kontak", "alamat",
		}),

		SystemPrompt: getEnv("SYSTEM_PROMPT", DefaultSystemPrompt),

		WAStoreDir: getEnv("WA_STORE_DIR", "./wa_stores"),

		CORSOrigins: getEnvSlice("CORS_ORIGINS", []string{"http://localhost:5173", "http://localhost:3000"}),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}

func getEnvSlice(key string, fallback []string) []string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		parts := strings.Split(value, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
		return result
	}
	return fallback
}

// Validate ensures required secrets and configuration values are present.
func (c Config) Validate() error {
	if strings.TrimSpace(c.JWTSecret) == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 bytes")
	}
	if strings.TrimSpace(c.EncryptionKey) == "" {
		return fmt.Errorf("ENCRYPTION_KEY is required")
	}
	if len(c.EncryptionKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes, got %d bytes", len(c.EncryptionKey))
	}
	if strings.TrimSpace(c.DatabaseURL) == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	return nil
}
