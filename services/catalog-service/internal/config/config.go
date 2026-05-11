package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCAddr   string        `envconfig:"CATALOG_GRPC_ADDR" default:":50053"`
	DBURL      string        `envconfig:"CATALOG_DB_URL" required:"true"`
	RedisAddr  string        `envconfig:"CATALOG_REDIS_ADDR" default:"redis:6379"`
	Brokers    []string      `envconfig:"CATALOG_KAFKA_BROKERS" default:"kafka:9092"`
	EventTopic string        `envconfig:"CATALOG_KAFKA_TOPIC" default:"catalog.events"`
	MigrateURL string        `envconfig:"CATALOG_MIGRATE_URL" default:"file:///app/migrations"`
	LogLevel   string        `envconfig:"CATALOG_LOG_LEVEL" default:"info"`
	CacheTTL   time.Duration `envconfig:"CATALOG_CACHE_TTL" default:"5m"`
}

func Load() (*Config, error) {
	c := &Config{}
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}
	return c, nil
}
