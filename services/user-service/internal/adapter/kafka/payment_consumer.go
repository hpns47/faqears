package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	pkgkafka "github.com/faqears/faqears/pkg/kafka"
	"github.com/faqears/faqears/services/user-service/internal/domain"
	"github.com/faqears/faqears/services/user-service/internal/usecase"
)

type PaymentEventsConsumer struct {
	c   *pkgkafka.Consumer
	uc  *usecase.User
	log *slog.Logger
}

type paymentFinishedEvent struct {
	UserID    string `json:"user_id"`
	PaymentID string `json:"payment_id"`
	Purpose   string `json:"purpose"`
	Status    string `json:"status"`
}

func NewPaymentEventsConsumer(brokers []string, topic, groupID string, uc *usecase.User, log *slog.Logger) *PaymentEventsConsumer {
	c := pkgkafka.NewConsumer(pkgkafka.ConsumerConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	}, log)
	return &PaymentEventsConsumer{c: c, uc: uc, log: log}
}

func (p *PaymentEventsConsumer) Run(ctx context.Context) error {
	return p.c.Run(ctx, p.handle)
}

func (p *PaymentEventsConsumer) handle(ctx context.Context, _, value []byte, headers map[string]string) error {
	if headers["event_type"] != "payment.finished" {
		return nil
	}
	var ev paymentFinishedEvent
	if err := json.Unmarshal(value, &ev); err != nil {
		return err
	}
	if ev.UserID == "" {
		return nil
	}
	p.log.InfoContext(ctx, "payment_finished received",
		slog.String("user_id", ev.UserID),
		slog.String("payment_id", ev.PaymentID),
		slog.String("purpose", ev.Purpose),
	)
	if ev.Purpose == "premium" || ev.Purpose == "" {
		if _, err := p.uc.UpdateTier(ctx, ev.UserID, domain.TierPremium); err != nil {
			return err
		}
	}
	return nil
}

func (p *PaymentEventsConsumer) Close() error {
	return p.c.Close()
}
