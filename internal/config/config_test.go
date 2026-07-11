package config

import (
	"context"
	"encoding/base64"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setEnv sets environment variables for a single test and restores them
// on cleanup.
func setEnv(t *testing.T, key, value string) {
	t.Helper()
	prev := os.Getenv(key)
	require.NoError(t, os.Setenv(key, value))
	t.Cleanup(func() { _ = os.Setenv(key, prev) })
}

func TestLoad_Defaults(t *testing.T) {
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	setEnv(t, "POLYGLOT_VAULT_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	setEnv(t, "POLYGLOT_JWT_SECRET", base64.StdEncoding.EncodeToString(make([]byte, 32)))

	cfg, err := Load(context.Background())
	require.NoError(t, err)

	assert.Equal(t, "development", cfg.Environment)
	assert.Equal(t, "http", cfg.MCPTransport)
	assert.Equal(t, ":8080", cfg.MCPHTTPAddr)
	assert.Equal(t, ":8081", cfg.RESTAddr)
	assert.Equal(t, 24*time.Hour, cfg.JWTExpiry)
}

func TestLoad_Production(t *testing.T) {
	setEnv(t, "DATABASE_URL", "postgres://localhost/prod")
	setEnv(t, "POLYGLOT_VAULT_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	setEnv(t, "POLYGLOT_JWT_SECRET", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	setEnv(t, "POLYGLOT_ENV", "production")
	setEnv(t, "POLYGLOT_JWT_EXPIRY", "1h")
	setEnv(t, "MCP_TRANSPORT", "stdio")
	setEnv(t, "MCP_HTTP_ADDR", ":3000")
	setEnv(t, "POLYGLOT_REST_ADDR", ":3001")

	cfg, err := Load(context.Background())
	require.NoError(t, err)

	assert.True(t, cfg.IsProduction())
	assert.Equal(t, time.Hour, cfg.JWTExpiry)
	assert.Equal(t, "stdio", cfg.MCPTransport)
	assert.Equal(t, ":3000", cfg.MCPHTTPAddr)
	assert.Equal(t, ":3001", cfg.RESTAddr)
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	setEnv(t, "DATABASE_URL", "")
	setEnv(t, "POLYGLOT_VAULT_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	setEnv(t, "POLYGLOT_JWT_SECRET", base64.StdEncoding.EncodeToString(make([]byte, 32)))

	_, err := Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL is required")
}

func TestLoad_InvalidVaultKey(t *testing.T) {
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	setEnv(t, "POLYGLOT_VAULT_KEY", base64.StdEncoding.EncodeToString(make([]byte, 16)))
	setEnv(t, "POLYGLOT_JWT_SECRET", base64.StdEncoding.EncodeToString(make([]byte, 32)))

	_, err := Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "POLYGLOT_VAULT_KEY")
}

func TestLoad_InvalidJWTSecret(t *testing.T) {
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	setEnv(t, "POLYGLOT_VAULT_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	setEnv(t, "POLYGLOT_JWT_SECRET", "short")

	_, err := Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "POLYGLOT_JWT_SECRET")
}

func TestLoad_InvalidMCPTransport(t *testing.T) {
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	setEnv(t, "POLYGLOT_VAULT_KEY", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	setEnv(t, "POLYGLOT_JWT_SECRET", base64.StdEncoding.EncodeToString(make([]byte, 32)))
	setEnv(t, "MCP_TRANSPORT", "grpc")

	_, err := Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MCP_TRANSPORT")
}

func TestLoad_InvalidBase64Key(t *testing.T) {
	setEnv(t, "DATABASE_URL", "postgres://localhost/test")
	setEnv(t, "POLYGLOT_VAULT_KEY", "not-valid-base64!!!")
	setEnv(t, "POLYGLOT_JWT_SECRET", base64.StdEncoding.EncodeToString(make([]byte, 32)))

	_, err := Load(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "POLYGLOT_VAULT_KEY")
}

func TestConfig_String_RedactsSecrets(t *testing.T) {
	cfg := &Config{
		Environment:  "development",
		DatabaseURL:  "postgres://localhost/test",
		VaultKey:     make([]byte, 32),
		JWTSecret:    make([]byte, 32),
		JWTExpiry:    24 * time.Hour,
		MCPTransport: "http",
		MCPHTTPAddr:  ":8080",
		RESTAddr:     ":8081",
	}

	s := cfg.String()
	assert.Contains(t, s, "VaultKeyLen=32")
	assert.Contains(t, s, "JWTSecretLen=32")
	assert.NotContains(t, s, string(cfg.VaultKey))
}
