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

func (p *Publisher) ProfileUpdated(ctx context.Context, userID, displayName string) error {
	msg := &eventsv1.ProfileUpdated{
		UserId:      userID,
		DisplayName: displayName,
		OccurredAt:  time.Now().UTC().Unix(),
	}
	value, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.p.Publish(ctx, []byte(userID), value, map[string]string{"event_type": "user.profile_updated"})
}

func (p *Publisher) UserFollowed(ctx context.Context, followerID, followeeID string) error {
	msg := &eventsv1.UserFollowed{
		FollowerId: followerID,
		FolloweeId: followeeID,
		OccurredAt: time.Now().UTC().Unix(),
	}
	value, err := proto.Marshal(msg)
	if err != nil {
		return err
	}
	return p.p.Publish(ctx, []byte(followerID), value, map[string]string{"event_type": "user.followed"})
}

func (p *Publisher) Close() error {
	return p.p.Close()
}
