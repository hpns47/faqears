package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	GRPCAddr   string        `envconfig:"AUTH_GRPC_ADDR" default:":50051"`
	DBURL      string        `envconfig:"AUTH_DB_URL" required:"true"`
	JWTSecret  string        `envconfig:"AUTH_JWT_SECRET" required:"true"`
	JWTIssuer  string        `envconfig:"AUTH_JWT_ISSUER" default:"faqears.auth"`
	AccessTTL  time.Duration `envconfig:"AUTH_JWT_ACCESS_TTL" default:"15m"`
	RefreshTTL time.Duration `envconfig:"AUTH_JWT_REFRESH_TTL" default:"720h"`
	BcryptCost int           `envconfig:"AUTH_BCRYPT_COST" default:"12"`
	Brokers    []string      `envconfig:"AUTH_KAFKA_BROKERS" default:"kafka:9092"`
	EventTopic string        `envconfig:"AUTH_KAFKA_TOPIC" default:"auth.events"`
	LogLevel   string        `envconfig:"AUTH_LOG_LEVEL" default:"info"`
	MigrateURL string        `envconfig:"AUTH_MIGRATE_URL" default:"file:///app/migrations"`
}

func Load() (*Config, error) {
	c := &Config{}
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}
	return c, nil
}
