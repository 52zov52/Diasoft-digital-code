package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv        string
	AppPort       string
	LogLevel      string
	Timeout       time.Duration
	JWTSecret     string
	APIKeyHR      string
	CORSOrigins   []string
	
	// Database
	DBHost        string
	DBPort        string
	DBName        string
	DBUser        string
	DBPass        string
	DBSSLMode     string
	DBMaxOpenConns int
	DBMaxIdleConns int
	DBConnMaxLifetime time.Duration
	
	// Redis ← ДОБАВИТЬ ЭТО
	RedisURL      string
	
	// Crypto
	AESKeyHex         string
	ED25519PrivateKeyHex string
	
	// Kafka
	KafkaBrokers    []string
}

func Load() (*Config, error) {
	_ = godotenv.Load() // ignores error if .env missing

	cfg := &Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		AppPort:     getEnv("APP_PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		Timeout:     getDuration("APP_TIMEOUT", 5*time.Second),
		JWTSecret:   "7f8a9b0c1d2e3f4a5b6c7d8e9f0a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a",
		APIKeyHR:    getEnv("HR_API_KEY", ""),
		CORSOrigins: splitEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"),
		
		// Database
		DBHost:            getEnv("DB_HOST", "localhost"),
		DBPort:            getEnv("DB_PORT", "5432"),
		DBName:            getEnv("DB_NAME", "diplomaverify"),
		DBUser:            getEnv("DB_USER", "app_user"),
		DBPass:            getEnv("DB_PASS", ""),
		DBSSLMode:         getEnv("DB_SSL_MODE", "require"),
		DBMaxOpenConns:    25,
		DBMaxIdleConns:    5,
		DBConnMaxLifetime: 30 * time.Minute,
		
		// Redis ← ДОБАВИТЬ ЭТО
		RedisURL: getEnv("REDIS_URL", "redis://localhost:6379/0"),
		
		// Crypto
		AESKeyHex:            "9a8b7c6d5e4f3a2b1c0d9e8f7a6b5c4d3e2f1a0b9c8d7e6f5a4b3c2d1e0f9a8b",
		ED25519PrivateKeyHex: "1a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f9a0b1c2d3e4f5a6b7c8d9e0f1a2b",
		
		// Kafka
		KafkaBrokers: splitEnv("KAFKA_BROKERS", "localhost:9092"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return fallback
		}
		return d
	}
	return fallback
}

func splitEnv(key, fallback string) []string {
	if v := os.Getenv(key); v != "" {
		// простой split по запятой, в prod лучше использовать strings.FieldsFunc
		return []string{v}
	}
	return []string{fallback}
}

func (c *Config) SlogLevel() slog.Level {
	switch c.LogLevel {
	case "debug": return slog.LevelDebug
	case "warn": return slog.LevelWarn
	case "error": return slog.LevelError
	default: return slog.LevelInfo
	}
}