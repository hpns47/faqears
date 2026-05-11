package kafka

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/segmentio/kafka-go"
)

type ConsumerConfig struct {
	Brokers []string
	GroupID string
	Topic   string
}

type Handler func(ctx context.Context, key, value []byte, headers map[string]string) error

type Consumer struct {
	r   *kafka.Reader
	log *slog.Logger
}

func NewConsumer(cfg ConsumerConfig, log *slog.Logger) *Consumer {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Brokers,
		GroupID:        cfg.GroupID,
		Topic:          cfg.Topic,
		MinBytes:       1,
		MaxBytes:       10 << 20,
		CommitInterval: 0,
	})
	return &Consumer{r: r, log: log}
}

func (c *Consumer) Run(ctx context.Context, h Handler) error {
	for {
		m, err := c.r.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		headers := make(map[string]string, len(m.Headers))
		for _, hh := range m.Headers {
			headers[hh.Key] = string(hh.Value)
		}
		if err := h(ctx, m.Key, m.Value, headers); err != nil {
			c.log.ErrorContext(ctx, "consumer handler failed",
				slog.String("topic", m.Topic),
				slog.Int("partition", m.Partition),
				slog.Int64("offset", m.Offset),
				slog.String("error", err.Error()),
			)
			continue
		}
		if err := c.r.CommitMessages(ctx, m); err != nil {
			c.log.ErrorContext(ctx, "commit failed", slog.String("error", err.Error()))
		}
	}
}

func (c *Consumer) Close() error {
	return c.r.Close()
}
