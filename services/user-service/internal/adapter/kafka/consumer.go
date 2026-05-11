package kafka

import (
	"context"
	"log/slog"

	"google.golang.org/protobuf/proto"

	eventsv1 "github.com/faqears/faqears/gen/go/events/v1"
	pkgkafka "github.com/faqears/faqears/pkg/kafka"
	"github.com/faqears/faqears/services/user-service/internal/usecase"
)

type AuthEventsConsumer struct {
	c   *pkgkafka.Consumer
	uc  *usecase.User
	log *slog.Logger
}

func NewAuthEventsConsumer(brokers []string, topic, groupID string, uc *usecase.User, log *slog.Logger) *AuthEventsConsumer {
	c := pkgkafka.NewConsumer(pkgkafka.ConsumerConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	}, log)
	return &AuthEventsConsumer{c: c, uc: uc, log: log}
}

func (a *AuthEventsConsumer) Run(ctx context.Context) error {
	return a.c.Run(ctx, a.handle)
}

func (a *AuthEventsConsumer) handle(ctx context.Context, _, value []byte, headers map[string]string) error {
	switch headers["event_type"] {
	case "auth.user_registered":
		var ev eventsv1.UserRegistered
		if err := proto.Unmarshal(value, &ev); err != nil {
			return err
		}
		a.log.InfoContext(ctx, "user_registered received",
			slog.String("user_id", ev.GetUserId()),
			slog.String("email", ev.GetEmail()),
		)
		return a.uc.Provision(ctx, ev.GetUserId(), ev.GetEmail())
	default:
		return nil
	}
}

func (a *AuthEventsConsumer) Close() error {
	return a.c.Close()
}
