package config

import (
	"testing"
	"time"
)

func TestLoadConfigCacheTTL(t *testing.T) {
	setRequiredEnvironment(t)

	tests := []struct {
		name    string
		value   string
		want    time.Duration
		wantErr bool
	}{
		{name: "valid", value: "30m", want: 30 * time.Minute},
		{name: "invalid", value: "later", wantErr: true},
		{name: "zero", value: "0s", wantErr: true},
		{name: "negative", value: "-1m", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CACHE_TTL", tt.value)
			cfg, err := LoadConfig()
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadConfig() error = %v, wantErr %t", err, tt.wantErr)
			}
			if !tt.wantErr && cfg.CacheConfig.TTL != tt.want {
				t.Errorf("CacheConfig.TTL = %v, want %v", cfg.CacheConfig.TTL, tt.want)
			}
		})
	}
}

func TestLoadConfigPorts(t *testing.T) {
	setRequiredEnvironment(t)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}
	tests := []struct {
		name string
		got  int
		want int
	}{
		{name: "database", got: cfg.DBConfig.Port, want: 5432},
		{name: "cache", got: cfg.CacheConfig.Port, want: 6379},
		{name: "application", got: cfg.Port, want: 8080},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("port = %d, want %d", tt.got, tt.want)
			}
		})
	}
}

func setRequiredEnvironment(t *testing.T) {
	t.Helper()
	values := map[string]string{
		"DB_HOST":     "localhost",
		"DB_PORT":     "5432",
		"DB_USER":     "postgres",
		"DB_PASSWORD": "password",
		"DB_NAME":     "url_shortener",
		"CACHE_HOST":  "localhost",
		"CACHE_PORT":  "6379",
		"APP_HOST":    "localhost",
		"APP_PORT":    "8080",
		"JWT_SECRET":  "test-secret",
		"PRODUCTION":  "false",
	}
	for key, value := range values {
		t.Setenv(key, value)
	}
}
