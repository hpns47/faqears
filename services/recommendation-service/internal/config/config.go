package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCAddr        string        `envconfig:"RECOMMENDATION_GRPC_ADDR" default:":50056"`
	DBURL           string        `envconfig:"RECOMMENDATION_DB_URL" required:"true"`
	RedisAddr       string        `envconfig:"RECOMMENDATION_REDIS_ADDR" default:"redis:6379"`
	Brokers         []string      `envconfig:"RECOMMENDATION_KAFKA_BROKERS" default:"kafka:9092"`
	PlayTopic       string        `envconfig:"RECOMMENDATION_KAFKA_PLAY_TOPIC" default:"play.events"`
	GroupID         string        `envconfig:"RECOMMENDATION_KAFKA_GROUP_ID" default:"recommendation-service"`
	RefreshInterval time.Duration `envconfig:"RECOMMENDATION_REFRESH_INTERVAL" default:"10m"`
	MigrateURL      string        `envconfig:"RECOMMENDATION_MIGRATE_URL" default:"file:///app/migrations"`
	LogLevel        string        `envconfig:"RECOMMENDATION_LOG_LEVEL" default:"info"`
}

func Load() (*Config, error) {
	c := &Config{}
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}
	return c, nil
}
