package config

import (
	"fmt"
	"os"
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
	AccessTTL   string
	RefreshTTL  string
	Production  bool
}

func LoadConfig() (*AppConfig, error) {
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
		AccessTTL:  getEnv("JWT_ACCESS_TOKEN_TTL"),
		RefreshTTL: getEnv("JWT_REFRESH_TOKEN_TTL"),
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
