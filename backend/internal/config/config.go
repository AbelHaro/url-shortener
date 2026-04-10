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

type AppConfig struct {
	DB         DBConfig
	Host       string
	Port       string
	JWTSecret  string
	Production bool
}

func LoadConfig() (*AppConfig, error) {
	cfg := &AppConfig{
		DB: DBConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			DBName:   getEnv("DB_NAME"),
		},
		Host:       getEnv("APP_HOST"),
		Port:       getEnv("APP_PORT"),
		JWTSecret:  getEnv("JWT_SECRET"),
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
		cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.DBName, sslMode)
}
