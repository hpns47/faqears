package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCAddr       string        `envconfig:"STREAMING_GRPC_ADDR" default:":50054"`
	DBURL          string        `envconfig:"STREAMING_DB_URL" required:"true"`
	RedisAddr      string        `envconfig:"STREAMING_REDIS_ADDR" default:"redis:6379"`
	Brokers        []string      `envconfig:"STREAMING_KAFKA_BROKERS" default:"kafka:9092"`
	KafkaTopic     string        `envconfig:"STREAMING_KAFKA_TOPIC" default:"play.events"`
	MinioEndpoint       string        `envconfig:"STREAMING_MINIO_ENDPOINT" default:"minio:9000"`
	MinioPublicEndpoint string        `envconfig:"STREAMING_MINIO_PUBLIC_ENDPOINT" default:"localhost:9000"`
	MinioAccessKey      string        `envconfig:"STREAMING_MINIO_ACCESS_KEY" default:"faqears"`
	MinioSecretKey      string        `envconfig:"STREAMING_MINIO_SECRET_KEY" default:"faqears-secret"`
	MinioUseSSL         bool          `envconfig:"STREAMING_MINIO_USE_SSL" default:"false"`
	MinioPublicUseSSL   bool          `envconfig:"STREAMING_MINIO_PUBLIC_USE_SSL" default:"false"`
	MinioRegion         string        `envconfig:"STREAMING_MINIO_REGION" default:"us-east-1"`
	MinioBucket         string        `envconfig:"STREAMING_MINIO_BUCKET" default:"audio-originals"`
	SignedURLTTL        time.Duration `envconfig:"STREAMING_SIGNED_URL_TTL" default:"15m"`
	MigrateURL          string        `envconfig:"STREAMING_MIGRATE_URL" default:"file:///app/migrations"`
	LogLevel            string        `envconfig:"STREAMING_LOG_LEVEL" default:"info"`
}

func Load() (*Config, error) {
	c := &Config{}
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}
	return c, nil
}
