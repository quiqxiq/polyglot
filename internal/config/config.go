package config

import (
	_ "embed"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/quixiq/polyglot/pkg/logger"
)

//go:embed rbac_model.conf
var RBACModelConf string

const (
	defaultEncryptionKey = "12345678901234567890123456789012"
	defaultJWTSecret     = "polyglot-super-secret-jwt-key-change-in-production"
	defaultDatabaseURL   = "postgres://postgres:postgres@localhost:5432/polyglot?sslmode=disable"
)

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
	UploadDir  string

	// Admin-only integrations. Empty values fail closed.
	MCPAdminToken string // POLYGLOT_MCP_TOKEN: full read/write MCP bearer token
	MCPReadToken  string // POLYGLOT_MCP_READ_TOKEN: read-only token for the harness
	WSToken       string // POLYGLOT_WS_TOKEN: optional static token for non-browser terminal clients

	// DeepSeek harness (admin-only NetOps and WhatsApp assistant).
	HarnessDSHPath        string // HARNESS_DSH_BIN: dsh executable name or path
	HarnessCSPatchPath    string // HARNESS_CS_PATCH: read-only customer-service overlay
	HarnessWorkDir        string // HARNESS_WORK_DIR: child process working directory
	HarnessMCPURL         string // POLYGLOT_MCP_URL: MCP endpoint exported to the harness
	HarnessTimeoutSeconds int    // HARNESS_TIMEOUT_SECONDS: one complete run
	HarnessMaxConcurrent  int    // HARNESS_MAX_CONCURRENT: simultaneous harness processes
	WAAssistantSessionID  uint   // WA_ASSISTANT_SESSION_ID: dedicated session; 0 disables

	// NetOps web UI (admin-only reverse proxy to the harness web app).
	HarnessWebURL      string // HARNESS_WEB_URL: loopback URL of `dsh web`
	HarnessWebToken    string // HARNESS_WEB_TOKEN: launch token printed by `dsh web`
	HarnessProxySecret string // HARNESS_PROXY_SECRET: HMAC key for NetOps cookies/header
	NetOpsSessionHours int    // NETOPS_SESSION_HOURS: browser session lifetime

	CORSOrigins []string

	// Scheduler ISP (fase 3).
	BillingCronSpec   string
	IsolationCronSpec string
	SchedulerEnabled  bool

	// Scheduler fase 4: pengirim WA & snapshot harian.
	WaSendCronSpec   string
	ReminderCronSpec string
	SnapshotCronSpec string

	// Scheduler Fase 4 (planning v2): worker antrean job router & drift check.
	RouterJobCronSpec string
	DriftCronSpec     string
	OltSyncCronSpec   string
	ACSMatchCronSpec  string
	ACSTelemetryCron  string
}

// Load reads infrastructure secrets and settings from environment variables or .env files.
func Load() Config {
	if err := godotenv.Load(); err != nil {
		logger.WithComponent("Config").Warn("No .env file found or error loading, using environment/default values")
	}
	_ = godotenv.Overload(".env.local")

	originsRaw := getEnv("CORS_ORIGINS", "http://localhost:5173,http://localhost:3000,http://127.0.0.1:5173,http://127.0.0.1:3000,http://localhost:3080,http://127.0.0.1:3080")
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

	port := getEnv("PORT", "8080")

	return Config{
		Port:                 port,
		AppEnv:               getEnv("APP_ENV", "development"),
		LogLevel:             getEnv("LOG_LEVEL", "info"),
		DatabaseURL:          getEnv("DATABASE_URL", defaultDatabaseURL),
		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379/0"),
		EncryptionKey:        getEnv("ENCRYPTION_KEY", defaultEncryptionKey),
		JWTSecret:            getEnv("JWT_SECRET", defaultJWTSecret),
		JWTExpiryHours:       getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		RefreshTokenTTLHours: getEnvAsInt("REFRESH_TOKEN_TTL_HOURS", 168),
		WAStoreDir:           getEnv("WA_STORE_DIR", "./data/whatsapp"),
		UploadDir:            getEnv("UPLOAD_DIR", "data/uploads"),
		MCPAdminToken:        getEnv("POLYGLOT_MCP_TOKEN", ""),
		MCPReadToken:         getEnv("POLYGLOT_MCP_READ_TOKEN", ""),
		WSToken:              getEnv("POLYGLOT_WS_TOKEN", ""),
		CORSOrigins:          origins,

		HarnessDSHPath:        getEnv("HARNESS_DSH_BIN", "dsh"),
		HarnessCSPatchPath:    getEnv("HARNESS_CS_PATCH", "deepseek-harness/packages/bundle/netops-app/netops-cs.patch.yml"),
		HarnessWorkDir:        getEnv("HARNESS_WORK_DIR", ""),
		HarnessMCPURL:         getEnv("POLYGLOT_MCP_URL", "http://127.0.0.1:"+port+"/mcp"),
		HarnessTimeoutSeconds: getEnvAsInt("HARNESS_TIMEOUT_SECONDS", 120),
		HarnessMaxConcurrent:  getEnvAsInt("HARNESS_MAX_CONCURRENT", 2),
		WAAssistantSessionID:  getEnvAsUint("WA_ASSISTANT_SESSION_ID", 0),

		HarnessWebURL:      getEnv("HARNESS_WEB_URL", "http://127.0.0.1:3080"),
		HarnessWebToken:    getEnv("HARNESS_WEB_TOKEN", ""),
		HarnessProxySecret: getEnv("HARNESS_PROXY_SECRET", ""),
		NetOpsSessionHours: getEnvAsInt("NETOPS_SESSION_HOURS", 8),

		// Scheduler ISP (fase 3). Kosongkan spesifikasi untuk menonaktifkan
		// job terkait.
		BillingCronSpec:   getEnv("BILLING_CRON", "0 6 * * *"),
		IsolationCronSpec: getEnv("ISOLATION_CRON", "@every 10m"),
		WaSendCronSpec:    getEnv("WA_SEND_CRON", "@every 30s"),
		ReminderCronSpec:  getEnv("REMINDER_CRON", "0 7 * * *"),
		SnapshotCronSpec:  getEnv("SNAPSHOT_CRON", "5 0 * * *"),
		RouterJobCronSpec: getEnv("ROUTER_JOB_CRON", "@every 1m"),
		DriftCronSpec:     getEnv("DRIFT_CRON", "0 3 * * *"),
		OltSyncCronSpec:   getEnv("OLT_SYNC_CRON", "@every 5m"),
		ACSMatchCronSpec:  getEnv("ACS_MATCH_CRON", "@every 15m"),
		ACSTelemetryCron:  getEnv("ACS_TELEMETRY_CRON", "@every 1m"),
		SchedulerEnabled:  getEnv("SCHEDULER_ENABLED", "true") == "true",
	}
}

// IsProduction reports whether the application runs with production safety
// checks enabled.
func (c Config) IsProduction() bool {
	switch strings.ToLower(strings.TrimSpace(c.AppEnv)) {
	case "prod", "production":
		return true
	default:
		return false
	}
}

// Validate enforces configuration invariants. Production additionally rejects
// the development defaults that would otherwise silently weaken security.
func (c *Config) Validate() error {
	if len(c.EncryptionKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes (got %d)", len(c.EncryptionKey))
	}

	minJWTLength := 16
	if c.IsProduction() {
		minJWTLength = 32
	}
	if len(c.JWTSecret) < minJWTLength {
		return fmt.Errorf("JWT_SECRET must be at least %d characters for security", minJWTLength)
	}

	if port, err := strconv.Atoi(strings.TrimSpace(c.Port)); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("PORT must be a valid TCP port (got %q)", c.Port)
	}
	if c.JWTExpiryHours <= 0 {
		return fmt.Errorf("JWT_EXPIRY_HOURS must be positive (got %d)", c.JWTExpiryHours)
	}
	if c.RefreshTokenTTLHours <= 0 {
		return fmt.Errorf("REFRESH_TOKEN_TTL_HOURS must be positive (got %d)", c.RefreshTokenTTLHours)
	}
	if c.HarnessTimeoutSeconds < 0 {
		return fmt.Errorf("HARNESS_TIMEOUT_SECONDS must not be negative (got %d)", c.HarnessTimeoutSeconds)
	}
	if c.HarnessMaxConcurrent < 0 {
		return fmt.Errorf("HARNESS_MAX_CONCURRENT must not be negative (got %d)", c.HarnessMaxConcurrent)
	}
	if c.NetOpsSessionHours <= 0 {
		return fmt.Errorf("NETOPS_SESSION_HOURS must be positive (got %d)", c.NetOpsSessionHours)
	}
	if c.NetOpsSessionHours > 24*7 {
		return fmt.Errorf("NETOPS_SESSION_HOURS must not exceed 168 (got %d)", c.NetOpsSessionHours)
	}
	if secret := strings.TrimSpace(c.HarnessProxySecret); secret != "" && len(secret) < 16 {
		return fmt.Errorf("HARNESS_PROXY_SECRET must be at least 16 characters")
	}
	if webURL := strings.TrimSpace(c.HarnessWebURL); webURL != "" {
		parsed, err := url.Parse(webURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
			return fmt.Errorf("HARNESS_WEB_URL must be an absolute http(s) URL")
		}
	}
	if strings.TrimSpace(c.HarnessWebToken) != "" && strings.TrimSpace(c.HarnessWebURL) == "" {
		return fmt.Errorf("HARNESS_WEB_URL must be set when HARNESS_WEB_TOKEN is set")
	}
	if c.MCPAdminToken != "" && c.MCPReadToken != "" && c.MCPAdminToken == c.MCPReadToken {
		return fmt.Errorf("POLYGLOT_MCP_READ_TOKEN must differ from POLYGLOT_MCP_TOKEN")
	}
	if c.WAAssistantSessionID > 0 {
		if strings.TrimSpace(c.HarnessDSHPath) == "" {
			return fmt.Errorf("HARNESS_DSH_BIN must be set when WA_ASSISTANT_SESSION_ID is enabled")
		}
		if strings.TrimSpace(c.HarnessCSPatchPath) == "" {
			return fmt.Errorf("HARNESS_CS_PATCH must be set when WA_ASSISTANT_SESSION_ID is enabled")
		}
		if strings.TrimSpace(c.HarnessMCPURL) == "" {
			return fmt.Errorf("POLYGLOT_MCP_URL must be set when WA_ASSISTANT_SESSION_ID is enabled")
		}
		if c.HarnessTimeoutSeconds <= 0 {
			return fmt.Errorf("HARNESS_TIMEOUT_SECONDS must be positive when WA_ASSISTANT_SESSION_ID is enabled")
		}
		if c.HarnessMaxConcurrent <= 0 {
			return fmt.Errorf("HARNESS_MAX_CONCURRENT must be positive when WA_ASSISTANT_SESSION_ID is enabled")
		}
		if strings.TrimSpace(c.MCPReadToken) == "" {
			return fmt.Errorf("POLYGLOT_MCP_READ_TOKEN must be set when WA_ASSISTANT_SESSION_ID is enabled")
		}
	}

	if !c.IsProduction() {
		return nil
	}

	if strings.TrimSpace(c.DatabaseURL) == "" || c.DatabaseURL == defaultDatabaseURL {
		return fmt.Errorf("DATABASE_URL must point to a production database, not the development default")
	}
	if c.JWTSecret == defaultJWTSecret {
		return fmt.Errorf("JWT_SECRET must not use the development default in production")
	}
	if c.EncryptionKey == defaultEncryptionKey {
		return fmt.Errorf("ENCRYPTION_KEY must not use the development default in production")
	}
	for _, origin := range c.CORSOrigins {
		if strings.TrimSpace(origin) == "*" {
			return fmt.Errorf("CORS_ORIGINS must not contain '*' in production")
		}
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

func getEnvAsUint(key string, defaultVal uint) uint {
	valStr := getEnv(key, "")
	if valStr == "" {
		return defaultVal
	}
	val, err := strconv.ParseUint(valStr, 10, 32)
	if err != nil {
		logger.WithComponent("Config").WithField("key", key).Warn("invalid unsigned integer; using default")
		return defaultVal
	}
	return uint(val)
}
