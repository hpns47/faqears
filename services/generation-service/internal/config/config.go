package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	GRPCAddr             string   `envconfig:"GENERATION_GRPC_ADDR" default:":50058"`
	DBURL                string   `envconfig:"GENERATION_DB_URL" required:"true"`
	Brokers              []string `envconfig:"GENERATION_KAFKA_BROKERS" default:"kafka:9092"`
	EventTopic           string   `envconfig:"GENERATION_KAFKA_TOPIC" default:"generation.events"`
	AnthropicAPIKey      string   `envconfig:"GENERATION_ANTHROPIC_API_KEY" default:""`
	AnthropicModel       string   `envconfig:"GENERATION_ANTHROPIC_MODEL" default:"claude-haiku-4-5-20251001"`
	SunoAPIKey           string   `envconfig:"GENERATION_SUNO_API_KEY" default:""`
	SunoBaseURL          string   `envconfig:"GENERATION_SUNO_BASE_URL" default:"https://api.suno.ai"`
	ReplicateAPIToken    string   `envconfig:"GENERATION_REPLICATE_API_TOKEN" default:""`
	UserGRPCAddr         string   `envconfig:"GENERATION_USER_GRPC_ADDR" default:"user-service:50052"`
	CatalogGRPCAddr      string   `envconfig:"GENERATION_CATALOG_GRPC_ADDR" default:"catalog-service:50053"`
	StreamingGRPCAddr    string   `envconfig:"GENERATION_STREAMING_GRPC_ADDR" default:"streaming-service:50054"`
	MigrateURL           string   `envconfig:"GENERATION_MIGRATE_URL" default:"file:///app/migrations"`
	LogLevel             string   `envconfig:"GENERATION_LOG_LEVEL" default:"info"`
}

func Load() (*Config, error) {
	c := &Config{}
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}
	return c, nil
}
