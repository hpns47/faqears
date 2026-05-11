package kafka

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
)

type ProducerConfig struct {
	Brokers []string
	Topic   string
}

type Producer struct {
	w *kafka.Writer
}

func NewProducer(cfg ProducerConfig) *Producer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Topic:                  cfg.Topic,
		Balancer:               &kafka.Hash{},
		RequiredAcks:           kafka.RequireAll,
		BatchTimeout:           50 * time.Millisecond,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}
	return &Producer{w: w}
}

func (p *Producer) Publish(ctx context.Context, key, value []byte, headers map[string]string) error {
	msg := kafka.Message{Key: key, Value: value}
	for k, v := range headers {
		msg.Headers = append(msg.Headers, kafka.Header{Key: k, Value: []byte(v)})
	}
	return p.w.WriteMessages(ctx, msg)
}

func (p *Producer) Close() error {
	return p.w.Close()
}
