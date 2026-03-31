package config

import (
	"fmt"
	"os"
)

type ServerConfig struct {
	Address     string
	Port        string
	Environment string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type Config struct {
	Server ServerConfig
	DB     DBConfig
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Address:     getEnv("SERVER_ADDRESS"),
			Port:        getEnv("SERVER_PORT"),
			Environment: getEnv("SERVER_ENV"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST"),
			Port:     getEnv("DB_PORT"),
			User:     getEnv("DB_USER"),
			Password: getEnv("DB_PASSWORD"),
			DBName:   getEnv("DB_NAME"),
			SSLMode:  getEnv("DB_SSL_MODE"),
		},
	}
}

func getEnv(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic(fmt.Sprintf("environment variable %s not set", key))
}

func (cfg *DBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}
