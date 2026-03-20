package util

import (
	"log"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBHost               string        `envconfig:"DB_HOST" default:"db"`
	DBPort               string        `envconfig:"DB_PORT" default:"3306"`
	DBUser               string        `envconfig:"DB_USER" default:"root"`
	DBPass               string        `envconfig:"DB_PASSWORD" default:"1234"`
	DBName               string        `envconfig:"DB_NAME" default:"spice_ledger"`
	JWTSecret            string        `envconfig:"JWT_SECRET" default:"supersecretjwtkey123!"`
	AccessTokenDuration  time.Duration `envconfig:"ACCESS_TOKEN_DURATION" default:"1m"`
	RefreshTokenDuration time.Duration `envconfig:"REFRESH_TOKEN_DURATION" default:"168h"`
	BasicAuthUser        string        `envconfig:"BASIC_AUTH_USER" default:"admin"`
	BasicAuthPass        string        `envconfig:"BASIC_AUTH_PASS" default:"secret123"`
	Port                 int           `envconfig:"PORT" default:"50051"`
	LogLevel             string        `envconfig:"LOG_LEVEL" default:"debug"`
}

func LoadConfig() *Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to process config: %v", err)
	}

	return &cfg
}
