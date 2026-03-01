package util

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBHost               string
	DBPort               string
	DBUser               string
	DBPass               string
	DBName               string
	JWTSecret            string
	AccessTokenDuration  time.Duration
	RefreshTokenDuration time.Duration
	BasicAuthUser        string
	BasicAuthPass        string
	Port                 int
	LogLevel             string
}

func LoadConfig() *Config {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	return &Config{
		DBHost:               getEnv("DB_HOST", "db"),
		DBPort:               getEnv("DB_PORT", "3306"),
		DBUser:               getEnv("DB_USER", "root"),
		DBPass:               getEnv("DB_PASSWORD", "1234"),
		DBName:               getEnv("DB_NAME", "spice_ledger"),
		JWTSecret:            getEnv("JWT_SECRET", "supersecretjwtkey123!"),
		AccessTokenDuration:  getDurationEnv("ACCESS_TOKEN_DURATION", 15*time.Minute),
		RefreshTokenDuration: getDurationEnv("REFRESH_TOKEN_DURATION", 168*time.Hour),
		BasicAuthUser:        getEnv("BASIC_AUTH_USER", "admin"),
		BasicAuthPass:        getEnv("BASIC_AUTH_PASS", "secret123"),
		Port:                 getIntEnv("PORT", 50051),
		LogLevel:             getEnv("LOG_LEVEL", "debug"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	valueStr := getEnv(key, "")
	if value, err := time.ParseDuration(valueStr); err == nil {
		return value
	}
	return defaultValue
}
