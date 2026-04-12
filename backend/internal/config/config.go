package config

import (
	"fmt"
	"os"
	"time"
)

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}
type RangeConfig struct {
	RangeSize   uint64
	RangeOffset uint64
}

type AppConfig struct {
	DBConfig    DBConfig
	RangeConfig RangeConfig
	Host        string
	Port        string
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

	cfg := &AppConfig{
		DBConfig: DBConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			DBName:   getEnv("DB_NAME"),
		},
		RangeConfig: RangeConfig{
			RangeSize:   1000,
			RangeOffset: 100,
		},
		Host:       getEnv("APP_HOST"),
		Port:       getEnv("APP_PORT"),
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

	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DBConfig.Host, cfg.DBConfig.Port, cfg.DBConfig.User, cfg.DBConfig.Password, cfg.DBConfig.DBName, sslMode)
}
