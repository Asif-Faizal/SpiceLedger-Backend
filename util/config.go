package util

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	AppEnv               string        `envconfig:"APP_ENV" default:"development"`
	DBHost               string        `envconfig:"DB_HOST" default:"db"`
	DBPort               string        `envconfig:"DB_PORT" default:"3306"`
	DBUser               string        `envconfig:"DB_USER" default:"root"`
	DBPass               string        `envconfig:"DB_PASSWORD" default:"1234"`
	DBName               string        `envconfig:"DB_NAME" default:"spice_ledger"`
	JWTSecret            string        `envconfig:"JWT_SECRET" default:"supersecretjwtkey123!"`
	AccessTokenDuration  time.Duration `envconfig:"ACCESS_TOKEN_DURATION" default:"30m"`
	RefreshTokenDuration time.Duration `envconfig:"REFRESH_TOKEN_DURATION" default:"168h"`
	BasicAuthUser        string        `envconfig:"BASIC_AUTH_USER" default:"admin"`
	BasicAuthPass        string        `envconfig:"BASIC_AUTH_PASS" default:"secret123"`
	ControlGrpcPort      int           `envconfig:"CONTROL_GRPC_PORT" default:"50051"`
	MarketGrpcPort       int           `envconfig:"MARKET_GRPC_PORT" default:"50052"`
	RestPort             int           `envconfig:"REST_PORT" default:"8082"`
	GraphqlPort          int           `envconfig:"GRAPHQL_PORT" default:"8081"`
	ProxyPort            int           `envconfig:"PROXY_PORT" default:"8080"`
	LogLevel             string        `envconfig:"LOG_LEVEL" default:"debug"`
	AccountGrpcURL       string        `envconfig:"ACCOUNT_GRPC_URL"`
	MarketGrpcURL        string        `envconfig:"MARKET_GRPC_URL"`
	AccountServiceURL    string        `envconfig:"ACCOUNT_SERVICE_URL"`
	GraphqlGatewayURL    string        `envconfig:"GRAPHQL_GATEWAY_URL"`
}

func LoadConfig() *Config {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		log.Fatalf("Failed to process config: %v", err)
	}

	cfg.Validate()
	return &cfg
}

func (c *Config) Validate() {
	if c.IsProduction() {
		if c.JWTSecret == "supersecretjwtkey123!" {
			log.Fatal("JWT_SECRET must be set to a strong value in production")
		}
		if c.DBPass == "1234" {
			log.Fatal("DB_PASSWORD must be changed from the default in production")
		}
		if c.BasicAuthPass == "secret123" {
			log.Fatal("BASIC_AUTH_PASS must be changed from the default in production")
		}
	}
}

func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.AppEnv, "production") || strings.EqualFold(c.AppEnv, "prod")
}

func (c *Config) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		c.DBUser, c.DBPass, c.DBHost, c.DBPort, c.DBName)
}

func (c *Config) ResolveAccountGrpcURL() string {
	if c.AccountGrpcURL != "" {
		return c.AccountGrpcURL
	}
	return fmt.Sprintf("localhost:%d", c.ControlGrpcPort)
}

func (c *Config) ResolveMarketGrpcURL() string {
	if c.MarketGrpcURL != "" {
		return c.MarketGrpcURL
	}
	return fmt.Sprintf("localhost:%d", c.MarketGrpcPort)
}

func (c *Config) ResolveAccountServiceURL() string {
	if c.AccountServiceURL != "" {
		return c.AccountServiceURL
	}
	return fmt.Sprintf("http://localhost:%d", c.RestPort)
}

func (c *Config) ResolveGraphqlGatewayURL() string {
	if c.GraphqlGatewayURL != "" {
		return c.GraphqlGatewayURL
	}
	return fmt.Sprintf("http://localhost:%d", c.GraphqlPort)
}
