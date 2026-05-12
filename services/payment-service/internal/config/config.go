package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	GRPCAddr       string `envconfig:"PAYMENT_GRPC_ADDR" default:":50057"`
	DBURL          string `envconfig:"PAYMENT_DB_URL" required:"true"`
	KafkaBrokers   []string `envconfig:"PAYMENT_KAFKA_BROKERS" default:"kafka:9092"`
	KafkaTopic     string `envconfig:"PAYMENT_KAFKA_TOPIC" default:"payment.events"`
	NowPaymentsBaseURL   string `envconfig:"PAYMENT_NOWPAYMENTS_BASE_URL" default:"https://api-sandbox.nowpayments.io"`
	NowPaymentsAPIKey    string `envconfig:"PAYMENT_NOWPAYMENTS_API_KEY" default:""`
	NowPaymentsIPNSecret string `envconfig:"PAYMENT_NOWPAYMENTS_IPN_SECRET" default:""`
	IPNCallbackURL string `envconfig:"PAYMENT_IPN_CALLBACK_URL" default:"http://localhost:8080/api/v1/payments/nowpayments/webhook"`
	MigrateURL     string `envconfig:"PAYMENT_MIGRATE_URL" default:"file:///app/migrations"`
	LogLevel       string `envconfig:"PAYMENT_LOG_LEVEL" default:"info"`
}

func Load() (*Config, error) {
	c := &Config{}
	if err := envconfig.Process("", c); err != nil {
		return nil, err
	}
	return c, nil
}
