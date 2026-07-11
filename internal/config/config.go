package config

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration for the Polyglot backend.
// Values are loaded from environment variables (and optional .env file)
// by Load.
type Config struct {
	// Environment is one of "development", "production", or "test".
	Environment string

	// DatabaseURL is the PostgreSQL connection string.
	DatabaseURL string

	// VaultKey is the 32-byte AES-GCM key used to encrypt/decrypt device
	// credentials. It must be exactly 32 bytes.
	VaultKey []byte

	// JWTSecret is the key used to sign and validate JWT tokens.
	// It should be at least 32 bytes.
	JWTSecret []byte

	// JWTExpiry is the lifetime of issued JWT tokens.
	JWTExpiry time.Duration

	// MCPTransport selects the MCP transport: "stdio" or "http".
	MCPTransport string

	// MCPHTTPAddr is the listen address for the MCP HTTP transport.
	MCPHTTPAddr string

	// RESTAddr is the listen address for the Gin REST API.
	RESTAddr string
}

// Load reads configuration from environment variables and an optional .env
// file. It validates required fields and returns an error if the
// configuration is invalid or incomplete.
func Load(ctx context.Context) (*Config, error) {
	// Best-effort .env load; ignore missing file so that container env-only
	// deployments still work.
	_ = godotenv.Load(".env")

	cfg := &Config{
		Environment:  envString("POLYGLOT_ENV", "development"),
		DatabaseURL:  os.Getenv("DATABASE_URL"),
		MCPTransport: envString("MCP_TRANSPORT", "http"),
		MCPHTTPAddr:  envString("MCP_HTTP_ADDR", ":8080"),
		RESTAddr:     envString("POLYGLOT_REST_ADDR", ":8081"),
	}

	vaultKey, err := decodeBase64Key("POLYGLOT_VAULT_KEY", 32)
	if err != nil {
		return nil, fmt.Errorf("config: vault key: %w", err)
	}
	cfg.VaultKey = vaultKey

	jwtSecret, err := decodeBase64Key("POLYGLOT_JWT_SECRET", 32)
	if err != nil {
		return nil, fmt.Errorf("config: jwt secret: %w", err)
	}
	cfg.JWTSecret = jwtSecret

	jwtExpiry, err := envDuration("POLYGLOT_JWT_EXPIRY", 24*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("config: jwt expiry: %w", err)
	}
	cfg.JWTExpiry = jwtExpiry

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// validate ensures required fields are present and well-formed.
func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("config: DATABASE_URL is required")
	}
	if len(c.VaultKey) != 32 {
		return fmt.Errorf("config: POLYGLOT_VAULT_KEY must decode to exactly 32 bytes, got %d", len(c.VaultKey))
	}
	if len(c.JWTSecret) < 32 {
		return fmt.Errorf("config: POLYGLOT_JWT_SECRET must decode to at least 32 bytes, got %d", len(c.JWTSecret))
	}
	if c.JWTExpiry <= 0 {
		return fmt.Errorf("config: POLYGLOT_JWT_EXPIRY must be positive")
	}
	if c.MCPTransport != "stdio" && c.MCPTransport != "http" {
		return fmt.Errorf("config: MCP_TRANSPORT must be 'stdio' or 'http', got %q", c.MCPTransport)
	}
	return nil
}

// envString returns the value of the environment variable key, or fallback
// if the variable is unset or empty.
func envString(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envDuration parses the environment variable key as a time.Duration,
// returning fallback if unset. It returns an error if the value is present
// but invalid.
func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", v, err)
	}
	return d, nil
}

// decodeBase64Key reads the environment variable key, decodes it from
// base64, and verifies it meets the minimum length. The value must be
// valid base64.
func decodeBase64Key(key string, minLen int) ([]byte, error) {
	v := os.Getenv(key)
	if v == "" {
		return nil, fmt.Errorf("%s is required", key)
	}

	decoded, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return nil, fmt.Errorf("%s must be valid base64: %w", key, err)
	}
	if len(decoded) < minLen {
		return nil, fmt.Errorf("%s decodes to %d bytes, need at least %d", key, len(decoded), minLen)
	}
	return decoded, nil
}

// IsProduction reports whether the environment is production.
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsTest reports whether the environment is test.
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}

// String returns a redacted configuration summary suitable for logs.
func (c *Config) String() string {
	return fmt.Sprintf(
		"Config{Environment=%s DatabaseURL=%s MCPTransport=%s MCPHTTPAddr=%s RESTAddr=%s VaultKeyLen=%d JWTSecretLen=%d JWTExpiry=%s}",
		c.Environment,
		c.DatabaseURL,
		c.MCPTransport,
		c.MCPHTTPAddr,
		c.RESTAddr,
		len(c.VaultKey),
		len(c.JWTSecret),
		c.JWTExpiry,
	)
}
