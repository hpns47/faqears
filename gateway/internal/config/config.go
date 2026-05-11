package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	HTTPAddr        string        `envconfig:"GATEWAY_HTTP_ADDR" default:":8080"`
	AuthGRPCAddr    string        `envconfig:"GATEWAY_AUTH_GRPC_ADDR" default:"auth-service:50051"`
	UserGRPCAddr    string        `envconfig:"GATEWAY_USER_GRPC_ADDR" default:"user-service:50052"`
	CatalogGRPCAddr string        `envconfig:"GATEWAY_CATALOG_GRPC_ADDR" default:"catalog-service:50053"`
	RedisAddr       string        `envconfig:"GATEWAY_REDIS_ADDR" default:"redis:6379"`
	TokenCacheTTL   time.Duration `envconfig:"GATEWAY_TOKEN_CACHE_TTL" default:"30s"`
	RateLimitAnon   int           `envconfig:"GATEWAY_RATE_LIMIT_ANON" default:"100"`
	RateLimitAuthed int           `envconfig:"GATEWAY_RATE_LIMIT_AUTHED" default:"1000"`
	LogLevel        string        `envconfig:"GATEWAY_LOG_LEVEL" default:"info"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
