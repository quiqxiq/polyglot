package config

import (
	"strings"
	"testing"
)

func validDevelopmentConfig() Config {
	return Config{
		Port:                 "8080",
		AppEnv:               "development",
		DatabaseURL:          "postgres://postgres:postgres@localhost:5432/polyglot?sslmode=disable",
		EncryptionKey:        strings.Repeat("e", 32),
		JWTSecret:            strings.Repeat("j", 16),
		JWTExpiryHours:       24,
		RefreshTokenTTLHours: 168,
		CORSOrigins:          []string{"http://localhost:5173"},
	}
}

func validProductionConfig() Config {
	cfg := validDevelopmentConfig()
	cfg.AppEnv = "production"
	cfg.DatabaseURL = "postgres://polyglot:secret@db.internal:5432/polyglot?sslmode=require"
	cfg.JWTSecret = strings.Repeat("j", 32)
	cfg.CORSOrigins = []string{"https://console.example.com"}
	return cfg
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*Config)
		wantErr string
	}{
		{
			name:   "valid development config",
			mutate: func(*Config) {},
		},
		{
			name:   "valid production config",
			mutate: func(c *Config) { *c = validProductionConfig() },
		},
		{
			name:    "encryption key must be 32 bytes",
			mutate:  func(c *Config) { c.EncryptionKey = "short" },
			wantErr: "ENCRYPTION_KEY",
		},
		{
			name:    "development JWT secret minimum",
			mutate:  func(c *Config) { c.JWTSecret = strings.Repeat("j", 15) },
			wantErr: "JWT_SECRET",
		},
		{
			name: "production JWT secret minimum",
			mutate: func(c *Config) {
				*c = validProductionConfig()
				c.JWTSecret = strings.Repeat("j", 16)
			},
			wantErr: "JWT_SECRET",
		},
		{
			name:    "invalid port",
			mutate:  func(c *Config) { c.Port = "not-a-port" },
			wantErr: "PORT",
		},
		{
			name:    "port out of range",
			mutate:  func(c *Config) { c.Port = "70000" },
			wantErr: "PORT",
		},
		{
			name:    "JWT expiry must be positive",
			mutate:  func(c *Config) { c.JWTExpiryHours = 0 },
			wantErr: "JWT_EXPIRY_HOURS",
		},
		{
			name:    "refresh TTL must be positive",
			mutate:  func(c *Config) { c.RefreshTokenTTLHours = 0 },
			wantErr: "REFRESH_TOKEN_TTL_HOURS",
		},
		{
			name: "production rejects default database",
			mutate: func(c *Config) {
				*c = validProductionConfig()
				c.DatabaseURL = defaultDatabaseURL
			},
			wantErr: "DATABASE_URL",
		},
		{
			name: "production rejects default JWT secret",
			mutate: func(c *Config) {
				*c = validProductionConfig()
				c.JWTSecret = defaultJWTSecret
			},
			wantErr: "JWT_SECRET",
		},
		{
			name: "production rejects default encryption key",
			mutate: func(c *Config) {
				*c = validProductionConfig()
				c.EncryptionKey = defaultEncryptionKey
			},
			wantErr: "ENCRYPTION_KEY",
		},
		{
			name: "production rejects wildcard CORS",
			mutate: func(c *Config) {
				*c = validProductionConfig()
				c.CORSOrigins = []string{"*"}
			},
			wantErr: "CORS_ORIGINS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validDevelopmentConfig()
			tt.mutate(&cfg)
			err := cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Validate() = nil, want error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("Validate() = %v, want error containing %q", err, tt.wantErr)
			}
		})
	}
}

func TestIsProduction(t *testing.T) {
	tests := []struct {
		env  string
		want bool
	}{
		{env: "production", want: true},
		{env: "PROD", want: true},
		{env: " production ", want: true},
		{env: "development", want: false},
		{env: "test", want: false},
		{env: "", want: false},
	}
	for _, tt := range tests {
		if got := (Config{AppEnv: tt.env}).IsProduction(); got != tt.want {
			t.Errorf("IsProduction(%q) = %v, want %v", tt.env, got, tt.want)
		}
	}
}
