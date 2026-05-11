package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	GRPCAddr        string   `envconfig:"USER_GRPC_ADDR" default:":50052"`
	DBURL           string   `envconfig:"USER_DB_URL" required:"true"`
	Brokers         []string `envconfig:"USER_KAFKA_BROKERS" default:"kafka:9092"`
	AuthTopic       string   `envconfig:"USER_KAFKA_AUTH_TOPIC" default:"auth.events"`
	EventTopic      string   `envconfig:"USER_KAFKA_TOPIC" default:"user.events"`
	GroupID         string   `envconfig:"USER_KAFKA_GROUP_ID" default:"user-service"`
	LogLevel        string   `envconfig:"USER_LOG_LEVEL" default:"info"`
	MigrateURL      string   `envconfig:"USER_MIGRATE_URL" default:"file:///app/migrations"`
}

func Load() (*Config, error) {
	c := &Config{}
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}
	return c, nil
}
