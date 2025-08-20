// internal/config/config.go
package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Nextcloud NextcloudConfig
	JWT       JWTConfig
	Environment string
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type NextcloudConfig struct {
	BaseURL       string
	AdminUser     string
	AdminPassword string
	APIVersion    string
}

type JWTConfig struct {
	SecretKey      string
	ExpirationHours int
}

func Load() (*Config, error) {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "localhost"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "postgresql:5432"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "casaos"),
			Password: getEnv("DB_PASSWORD", "casaos"),
			Name:     getEnv("DB_NAME", "casaos"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Nextcloud: NextcloudConfig{
			BaseURL:       getEnv("NEXTCLOUD_URL", "http://10.0.2.15:10081"),
			AdminUser:     getEnv("NEXTCLOUD_ADMIN_USER", "dvelazquez"),
			AdminPassword: getEnv("NEXTCLOUD_ADMIN_PASSWORD", "casaos"),
			APIVersion:    getEnv("NEXTCLOUD_API_VERSION", "v2"),
		},
		JWT: JWTConfig{
			SecretKey:       getEnv("JWT_SECRET", "Encelad0_123"),
			ExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		},
		Environment: getEnv("ENVIRONMENT", "development"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}