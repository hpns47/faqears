package kafka

import (
	"context"
	"encoding/json"
	"time"

	pkgkafka "github.com/faqears/faqears/pkg/kafka"
	"github.com/faqears/faqears/services/payment-service/internal/domain"
)

type Publisher struct {
	p *pkgkafka.Producer
}

func NewPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{
		p: pkgkafka.NewProducer(pkgkafka.ProducerConfig{Brokers: brokers, Topic: topic}),
	}
}

type eventPayload struct {
	PaymentID string `json:"payment_id"`
	UserID    string `json:"user_id"`
	Purpose   string `json:"purpose"`
	Status    string `json:"status"`
	OccurredAt int64 `json:"occurred_at"`
}

func (pub *Publisher) publish(ctx context.Context, eventType string, p *domain.Payment) error {
	payload := eventPayload{
		PaymentID:  p.ID,
		UserID:     p.UserID,
		Purpose:    p.Purpose,
		Status:     p.Status,
		OccurredAt: time.Now().UTC().Unix(),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": eventType}
	return pub.p.Publish(ctx, []byte(p.ID), b, headers)
}

func (pub *Publisher) PaymentInvoiceCreated(ctx context.Context, p *domain.Payment) error {
	return pub.publish(ctx, "payment.invoice_created", p)
}

func (pub *Publisher) PaymentStatusChanged(ctx context.Context, p *domain.Payment) error {
	return pub.publish(ctx, "payment.status_changed", p)
}

func (pub *Publisher) PaymentFinished(ctx context.Context, p *domain.Payment) error {
	return pub.publish(ctx, "payment.finished", p)
}

func (pub *Publisher) PaymentFailed(ctx context.Context, p *domain.Payment) error {
	return pub.publish(ctx, "payment.failed", p)
}

func (pub *Publisher) Close() error {
	return pub.p.Close()
}
