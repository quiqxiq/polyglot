package config

import (
	_ "embed"
	"fmt"
	"github.com/quixiq/polyglot/pkg/logger"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

//go:embed rbac_model.conf
var RBACModelConf string

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

	SessionTimeoutMinutes int
	SlidingWindowSize     int

	RateLimitPerMinute int
	RateLimitPerHour   int

	LLMMaxOutputTokens int

	AllowedTopics []string
	SystemPrompt  string

	WAStoreDir string

	CORSOrigins []string

	// WhatsApp Bot Multi-Tier Rate Limiter
	BotBurstLimit        int
	BotBurstWindowSecs   int
	BotMute1HourSecs     int
	BotBan24HourSecs     int
	BotDailyChatLimit    int
	BotWhitelistPhones   []string
}

func Load() Config {
	if err := godotenv.Load(); err != nil {
		logger.WithComponent("Config").Warn("No .env file found or error loading, using environment/default values")
	}
	_ = godotenv.Overload(".env.local")

	prompt := os.Getenv("SYSTEM_PROMPT")

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

	whitelistRaw := os.Getenv("BOT_WHITELIST_PHONES")
	var whitelist []string
	for _, p := range strings.Split(whitelistRaw, ",") {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			whitelist = append(whitelist, trimmed)
		}
	}

	return Config{
		Port:                  getEnv("PORT", "8080"),
		AppEnv:                getEnv("APP_ENV", "development"),
		LogLevel:              getEnv("LOG_LEVEL", "info"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/polyglot?sslmode=disable"),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379/0"),
		EncryptionKey:         getEnv("ENCRYPTION_KEY", "12345678901234567890123456789012"),
		JWTSecret:             getEnv("JWT_SECRET", "polyglot-super-secret-jwt-key-change-in-production"),
		JWTExpiryHours:        getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		RefreshTokenTTLHours:  getEnvAsInt("REFRESH_TOKEN_TTL_HOURS", 168),
		SessionTimeoutMinutes: getEnvAsInt("SESSION_TIMEOUT_MINUTES", 30),
		SlidingWindowSize:     getEnvAsInt("SLIDING_WINDOW_SIZE", 10),
		RateLimitPerMinute:    getEnvAsInt("RATE_LIMIT_PER_MINUTE", 60),
		RateLimitPerHour:      getEnvAsInt("RATE_LIMIT_PER_HOUR", 1000),
		LLMMaxOutputTokens:    getEnvAsInt("LLM_MAX_OUTPUT_TOKENS", 1024),
		AllowedTopics:         strings.Split(getEnv("ALLOWED_TOPICS", "hotspot,billing,network,general"), ","),
		SystemPrompt:          prompt,
		WAStoreDir:            getEnv("WA_STORE_DIR", "./data/whatsapp"),
		CORSOrigins:           origins,
		BotBurstLimit:         getEnvAsInt("BOT_BURST_LIMIT", 3),
		BotBurstWindowSecs:    getEnvAsInt("BOT_BURST_WINDOW_SECS", 5),
		BotMute1HourSecs:      getEnvAsInt("BOT_MUTE_1H_SECS", 3600),
		BotBan24HourSecs:      getEnvAsInt("BOT_BAN_24H_SECS", 86400),
		BotDailyChatLimit:     getEnvAsInt("BOT_DAILY_CHAT_LIMIT", 10),
		BotWhitelistPhones:    whitelist,
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
