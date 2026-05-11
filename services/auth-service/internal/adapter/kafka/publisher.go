package kafka

import (
	"context"
	"time"

	"google.golang.org/protobuf/proto"

	eventsv1 "github.com/faqears/faqears/gen/go/events/v1"
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

func (p *Publisher) UserRegistered(ctx context.Context, userID, email string) error {
	msg := &eventsv1.UserRegistered{
		UserId:     userID,
		Email:      email,
		OccurredAt: time.Now().UTC().Unix(),
	}
	value, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": "auth.user_registered"}
	return p.p.Publish(ctx, []byte(userID), value, headers)
}

func (p *Publisher) UserLoggedIn(ctx context.Context, userID string) error {
	msg := &eventsv1.UserLoggedIn{
		UserId:     userID,
		OccurredAt: time.Now().UTC().Unix(),
	}
	value, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": "auth.user_logged_in"}
	return p.p.Publish(ctx, []byte(userID), value, headers)
}

func (p *Publisher) Close() error {
	return p.p.Close()
}
