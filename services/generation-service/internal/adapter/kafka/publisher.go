package kafka

import (
	"context"
	"encoding/json"
	"time"

	pkgkafka "github.com/faqears/faqears/pkg/kafka"
)

type Publisher struct {
	p *pkgkafka.Producer
}

func NewPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{
		p: pkgkafka.NewProducer(pkgkafka.ProducerConfig{Brokers: brokers, Topic: topic}),
	}
}

type event struct {
	EventType  string `json:"event_type"`
	JobID      string `json:"job_id"`
	UserID     string `json:"user_id"`
	OccurredAt int64  `json:"occurred_at"`
}

func (p *Publisher) Publish(ctx context.Context, eventType, jobID, userID string) error {
	e := event{
		EventType:  eventType,
		JobID:      jobID,
		UserID:     userID,
		OccurredAt: time.Now().UTC().Unix(),
	}
	value, err := json.Marshal(e)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": eventType}
	return p.p.Publish(ctx, []byte(jobID), value, headers)
}

func (p *Publisher) Close() error {
	return p.p.Close()
}
