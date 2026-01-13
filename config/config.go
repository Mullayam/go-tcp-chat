package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

// Config holds all application configuration
type Config struct {
	// TCP Server
	TCPPort string

	// SMTP Configuration
	SMTPHost     string
	SMTPPort     int
	SMTPEmail    string
	SMTPPassword string

	// OTP Settings
	OTPExpirationMinutes int
	OTPMaxRetries        int

	// Username Validation
	UsernameMinLength int
	UsernameMaxLength int
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		TCPPort:              getEnv("TCP_PORT", "8888"),
		SMTPHost:             getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:             getEnvAsInt("SMTP_PORT", 587),
		SMTPEmail:            getEnv("SMTP_EMAIL", ""),
		SMTPPassword:         getEnv("SMTP_PASSWORD", ""),
		OTPExpirationMinutes: getEnvAsInt("OTP_EXPIRATION_MINUTES", 5),
		OTPMaxRetries:        getEnvAsInt("OTP_MAX_RETRIES", 3),
		UsernameMinLength:    getEnvAsInt("USERNAME_MIN_LENGTH", 3),
		UsernameMaxLength:    getEnvAsInt("USERNAME_MAX_LENGTH", 16),
	}

	// Validate required fields
	if cfg.SMTPEmail == "" {
		return nil, fmt.Errorf("SMTP_EMAIL is required")
	}
	if cfg.SMTPPassword == "" {
		return nil, fmt.Errorf("SMTP_PASSWORD is required")
	}

	return cfg, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(strings.TrimSpace(valueStr))
	if err != nil {
		return defaultValue
	}
	return value
}
