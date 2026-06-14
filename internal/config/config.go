package config

import (
	"fmt"
	"os"
)

// Config содержит все настройки приложения, считанные из переменных окружения.
type Config struct {
	DatabaseURL string
	JWTSecret   string
	ServerPort  string
	LogLevel    string

	// PGP-ключ для шифрования данных карт
	PGPPublicKey  string
	PGPPrivateKey string
	PGPPassphrase string

	// HMAC-секрет для подписи данных карт
	HMACSecret string

	// SMTP-настройки для email-уведомлений
	SMTPHost     string
	SMTPPort     string
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
}

// Load читает конфиг из переменных окружения. Обязательные поля проверяются явно.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		ServerPort:    os.Getenv("SERVER_PORT"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		PGPPublicKey:  os.Getenv("PGP_PUBLIC_KEY"),
		PGPPrivateKey: os.Getenv("PGP_PRIVATE_KEY"),
		PGPPassphrase: os.Getenv("PGP_PASSPHRASE"),
		HMACSecret:    getEnv("HMAC_SECRET", "default-hmac-secret-change-me"),
		SMTPHost:      getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:      getEnv("SMTP_PORT", "587"),
		SMTPUser:      os.Getenv("SMTP_USER"),
		SMTPPassword:  os.Getenv("SMTP_PASSWORD"),
		SMTPFrom:      getEnv("SMTP_FROM", os.Getenv("SMTP_USER")),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	if cfg.ServerPort == "" {
		return nil, fmt.Errorf("SERVER_PORT is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}
