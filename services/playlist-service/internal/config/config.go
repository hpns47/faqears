package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCAddr     string        `envconfig:"PLAYLIST_GRPC_ADDR" default:":50055"`
	DBURL        string        `envconfig:"PLAYLIST_DB_URL" required:"true"`
	RedisAddr    string        `envconfig:"PLAYLIST_REDIS_ADDR" default:"redis:6379"`
	Brokers      []string      `envconfig:"PLAYLIST_KAFKA_BROKERS" default:"kafka:9092"`
	EventTopic   string        `envconfig:"PLAYLIST_KAFKA_TOPIC" default:"playlist.events"`
	CatalogTopic string        `envconfig:"PLAYLIST_KAFKA_CATALOG_TOPIC" default:"catalog.events"`
	GroupID      string        `envconfig:"PLAYLIST_KAFKA_GROUP_ID" default:"playlist-service"`
	CacheTTL     time.Duration `envconfig:"PLAYLIST_CACHE_TTL" default:"5m"`
	MigrateURL   string        `envconfig:"PLAYLIST_MIGRATE_URL" default:"file:///app/migrations"`
	LogLevel     string        `envconfig:"PLAYLIST_LOG_LEVEL" default:"info"`
}

func Load() (*Config, error) {
	c := &Config{}
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}
	return c, nil
}
