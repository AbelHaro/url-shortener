package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}
type RangeConfig struct {
	RangeSize   uint64
	RangeOffset uint64
}

// CacheConfig contains the Valkey connection and URL-entry lifetime.
type CacheConfig struct {
	Host string
	Port int
	TTL  time.Duration
}
type AppConfig struct {
	DBConfig    DBConfig
	RangeConfig RangeConfig
	CacheConfig CacheConfig
	Host        string
	Port        int
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
	Production  bool
}

func LoadConfig() (*AppConfig, error) {
	accessTTLStr := getEnvOrDefault("JWT_ACCESS_TOKEN_TTL", "15m")
	accessTTL, err := time.ParseDuration(accessTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_TOKEN_TTL format: %w", err)
	}

	refreshTTLStr := getEnvOrDefault("JWT_REFRESH_TOKEN_TTL", "168h")
	refreshTTL, err := time.ParseDuration(refreshTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_TOKEN_TTL format: %w", err)
	}

	cacheTTLStr := getEnvOrDefault("CACHE_TTL", "1h")
	cacheTTL, err := time.ParseDuration(cacheTTLStr)
	if err != nil {
		return nil, fmt.Errorf("invalid CACHE_TTL format: %w", err)
	}
	if cacheTTL <= 0 {
		return nil, fmt.Errorf("CACHE_TTL must be positive")
	}

	cfg := &AppConfig{
		DBConfig: DBConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnvAsInt("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			DBName:   getEnv("DB_NAME"),
		},
		RangeConfig: RangeConfig{
			RangeSize:   1000,
			RangeOffset: 100,
		},
		CacheConfig: CacheConfig{
			Host: getEnv("CACHE_HOST"),
			Port: getEnvAsInt("CACHE_PORT"),
			TTL:  cacheTTL,
		},
		Host:       getEnv("APP_HOST"),
		Port:       getEnvAsInt("APP_PORT"),
		JWTSecret:  getEnv("JWT_SECRET"),
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
		Production: getEnv("PRODUCTION") == "true",
	}
	return cfg, nil
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic(fmt.Sprintf("environment variable %s not set", key))
}

func getEnvAsInt(key string) int {
	if value := os.Getenv(key); value != "" {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			panic(fmt.Sprintf("environment variable %s is not a valid integer: %v", key, err))
		}
		return intValue
	}
	panic(fmt.Sprintf("environment variable %s not set", key))
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func (cfg *AppConfig) DSN() string {

	var sslMode string
	if cfg.Production {
		sslMode = "require"
	} else {
		sslMode = "disable"
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBConfig.Host, cfg.DBConfig.Port, cfg.DBConfig.User, cfg.DBConfig.Password, cfg.DBConfig.DBName, sslMode)
}
