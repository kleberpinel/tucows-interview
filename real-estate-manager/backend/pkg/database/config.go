package database

import "os"

func NewConfigFromEnv() Config {
    return Config{
        Host:     getEnvOrDefault("DB_HOST", "localhost"),
        Port:     getEnvOrDefault("DB_PORT", "3306"),
        User:     getEnvOrDefault("DB_USER", "appuser"),
        Password: getEnvOrDefault("DB_PASSWORD", "apppassword"),
        DBName:   getEnvOrDefault("DB_NAME", "real_estate_db"),
    }
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}