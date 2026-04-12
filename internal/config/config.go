package config

import (
	"errors"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Env             string `env:"TODO_ENV" envDefault:"dev"`
	Port            int    `env:"PORT" envDefault:"80"`
	DBHost          string `env:"TODO_DB_HOST" envDefault:"127.0.0.1"`
	DBPort          int    `env:"TODO_DB_PORT" envDefault:"33306"`
	DBUser          string `env:"TODO_DB_USER" envDefault:"todo"`
	DBPassword      string `env:"TODO_DB_PASSWORD" envDefault:"todo"`
	DBName          string `env:"TODO_DB_NAME" envDefault:"todo"`
	JWTSecret       string `env:"TODO_JWT_SECRET" envDefault:"dev-secret-change-me"`
	AccessTokenTTL  int    `env:"TODO_ACCESS_TOKEN_TTL" envDefault:"900"`
	RefreshTokenTTL int    `env:"TODO_REFRESH_TOKEN_TTL" envDefault:"604800"`
	MetricsPort     int    `env:"METRICS_PORT" envDefault:"9090"`
	OTLPEndpoint    string `env:"OTEL_EXPORTER_OTLP_ENDPOINT" envDefault:""`
	ServiceName     string `env:"OTEL_SERVICE_NAME" envDefault:"todo-api"`
	CORSOrigin      string `env:"CORS_ORIGIN" envDefault:"http://localhost:3000"`
	SeedPassword    string `env:"SEED_PASSWORD" envDefault:"Seedpass12345!"`
}

func New() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if cfg.Env != "dev" && cfg.JWTSecret == "dev-secret-change-me" {
		return nil, errors.New("TODO_JWT_SECRET must be set in non-dev environment")
	}
	return cfg, nil
}
