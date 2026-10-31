package config

import (
	_ "embed"
	"encoding/base64"
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
	UploadDir  string

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

	// AI Router Ops (PLAN-AI-ROUTER-OPS v2).
	AIRouterOpsEnabled            bool
	AIV6Mutations                 bool
	AIMaxProposalsPerMinute       int
	AIMaxScriptLines              int
	AIChatStreamEnabled           bool
	AIChatMaxContextMessages      int
	AIChatEventRetentionDays      int
	CommandLogRetentionDays       int
	ApprovalTTLMinutes            int
	ApprovalRequireReasonOnReject bool
	RestorePointDir               string
	RestorePointRetentionDays     int
	RestorePointKeepLast          int
	RestorePointEncryptionKey     string
	RestorePointMaxBytes          int64
	RestorePointDedupe            bool
	BackupDailyEnabled            bool
	BackupDailyAt                 string
	BackupWeeklyEnabled           bool
	BackupWeeklyDay               string
	BackupWeeklyAt                string
	WatchdogEnabled               bool
	WatchdogRollbackSeconds       int
	ActiveDeviceTTLHours          int
	AISingleDeviceFallback        bool
	InspectTimeoutSeconds         int
	InspectCacheTTLMinutes        int
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
		UploadDir:            getEnv("UPLOAD_DIR", "data/uploads"),
		CORSOrigins:          origins,

		// Scheduler ISP (fase 3). Kosongkan spesifikasi untuk menonaktifkan
		// job terkait.
		BillingCronSpec:   getEnv("BILLING_CRON", "0 6 * * *"),
		IsolationCronSpec: getEnv("ISOLATION_CRON", "@every 10m"),
		WaSendCronSpec:    getEnv("WA_SEND_CRON", "@every 30s"),
		ReminderCronSpec:  getEnv("REMINDER_CRON", "0 7 * * *"),
		SnapshotCronSpec:  getEnv("SNAPSHOT_CRON", "5 0 * * *"),
		RouterJobCronSpec: getEnv("ROUTER_JOB_CRON", "@every 1m"),
		DriftCronSpec:     getEnv("DRIFT_CRON", "0 3 * * *"),
		SchedulerEnabled:  getEnv("SCHEDULER_ENABLED", "true") == "true",

		// AI Router Ops (PLAN-AI-ROUTER-OPS v2). Default aman: fitur aktif
		// tetapi mutation AI selalu butuh approval; watchdog OFF.
		AIRouterOpsEnabled:            getEnvAsBool("AI_ROUTER_OPS_ENABLED", true),
		AIV6Mutations:                 getEnvAsBool("AI_V6_MUTATIONS", true),
		AIMaxProposalsPerMinute:       getEnvAsInt("AI_MAX_PROPOSALS_PER_MINUTE", 10),
		AIMaxScriptLines:              getEnvAsInt("AI_MAX_SCRIPT_LINES", 50),
		AIChatStreamEnabled:           getEnvAsBool("AI_CHAT_STREAM_ENABLED", true),
		AIChatMaxContextMessages:      getEnvAsInt("AI_CHAT_MAX_CONTEXT_MESSAGES", 50),
		AIChatEventRetentionDays:      getEnvAsInt("AI_CHAT_EVENT_RETENTION_DAYS", 30),
		CommandLogRetentionDays:       getEnvAsInt("COMMAND_LOG_RETENTION_DAYS", 180),
		ApprovalTTLMinutes:            getEnvAsInt("APPROVAL_TTL_MINUTES", 15),
		ApprovalRequireReasonOnReject: getEnvAsBool("APPROVAL_REQUIRE_REASON_ON_REJECT", true),
		RestorePointDir:               getEnv("RESTORE_POINT_DIR", "data/restorepoints"),
		RestorePointRetentionDays:     getEnvAsInt("RESTORE_POINT_RETENTION_DAYS", 14),
		RestorePointKeepLast:          getEnvAsInt("RESTORE_POINT_KEEP_LAST", 30),
		RestorePointEncryptionKey:     getEnv("RESTORE_POINT_ENCRYPTION_KEY", ""),
		RestorePointMaxBytes:          int64(getEnvAsInt("RESTORE_POINT_MAX_BYTES", 52428800)),
		RestorePointDedupe:            getEnvAsBool("RESTORE_POINT_DEDUPE", true),
		BackupDailyEnabled:            getEnvAsBool("BACKUP_DAILY_ENABLED", true),
		BackupDailyAt:                 getEnv("BACKUP_DAILY_AT", "02:00"),
		BackupWeeklyEnabled:           getEnvAsBool("BACKUP_WEEKLY_ENABLED", true),
		BackupWeeklyDay:               getEnv("BACKUP_WEEKLY_DAY", "sun"),
		BackupWeeklyAt:                getEnv("BACKUP_WEEKLY_AT", "03:00"),
		WatchdogEnabled:               getEnvAsBool("WATCHDOG_ENABLED", false),
		WatchdogRollbackSeconds:       getEnvAsInt("WATCHDOG_ROLLBACK_SECONDS", 600),
		ActiveDeviceTTLHours:          getEnvAsInt("ACTIVE_DEVICE_TTL_HOURS", 24),
		AISingleDeviceFallback:        getEnvAsBool("AI_SINGLE_DEVICE_FALLBACK", false),
		InspectTimeoutSeconds:         getEnvAsInt("INSPECT_TIMEOUT_SECONDS", 5),
		InspectCacheTTLMinutes:        getEnvAsInt("INSPECT_CACHE_TTL_MINUTES", 10),
	}
}

func (c *Config) Validate() error {
	if len(c.EncryptionKey) != 32 {
		return fmt.Errorf("ENCRYPTION_KEY must be exactly 32 bytes (got %d)", len(c.EncryptionKey))
	}
	if len(c.JWTSecret) < 16 {
		return fmt.Errorf("JWT_SECRET must be at least 16 characters for security")
	}
	if c.ApprovalTTLMinutes <= 0 {
		return fmt.Errorf("APPROVAL_TTL_MINUTES must be greater than zero")
	}
	if c.RestorePointKeepLast <= 0 {
		return fmt.Errorf("RESTORE_POINT_KEEP_LAST must be greater than zero")
	}
	if c.AIRouterOpsEnabled && strings.EqualFold(c.AppEnv, "production") {
		key := c.RestorePointEncryptionKey
		if key == "" {
			return fmt.Errorf("RESTORE_POINT_ENCRYPTION_KEY is required in production when AI_ROUTER_OPS_ENABLED=true")
		}
		if len(key) != 32 {
			if decoded, err := base64.StdEncoding.DecodeString(key); err != nil || len(decoded) != 32 {
				return fmt.Errorf("RESTORE_POINT_ENCRYPTION_KEY must be a raw 32-byte string or base64-encoded 32 bytes")
			}
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

func getEnvAsBool(key string, defaultVal bool) bool {
	valStr := strings.ToLower(strings.TrimSpace(getEnv(key, "")))
	if valStr == "" {
		return defaultVal
	}
	return valStr == "true" || valStr == "1" || valStr == "yes" || valStr == "on"
}
