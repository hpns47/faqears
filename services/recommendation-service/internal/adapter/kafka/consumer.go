package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	pkgkafka "github.com/faqears/faqears/pkg/kafka"
	"github.com/faqears/faqears/services/recommendation-service/internal/domain"
	"github.com/faqears/faqears/services/recommendation-service/internal/port"
)

type playEventMsg struct {
	EventType  string `json:"event_type"`
	UserID     string `json:"user_id"`
	TrackID    string `json:"track_id"`
	SessionID  string `json:"session_id"`
	OccurredAt int64  `json:"occurred_at"`
}

type PlayEventsConsumer struct {
	c    *pkgkafka.Consumer
	repo port.PlayEventRepository
	log  *slog.Logger
}

func NewPlayEventsConsumer(brokers []string, topic, groupID string, repo port.PlayEventRepository, log *slog.Logger) *PlayEventsConsumer {
	c := pkgkafka.NewConsumer(pkgkafka.ConsumerConfig{
		Brokers: brokers,
		GroupID: groupID,
		Topic:   topic,
	}, log)
	return &PlayEventsConsumer{c: c, repo: repo, log: log}
}

func (p *PlayEventsConsumer) Run(ctx context.Context) error {
	return p.c.Run(ctx, p.handle)
}

func (p *PlayEventsConsumer) handle(ctx context.Context, _, value []byte, _ map[string]string) error {
	var msg playEventMsg
	if err := json.Unmarshal(value, &msg); err != nil {
		p.log.WarnContext(ctx, "play event unmarshal failed", slog.String("error", err.Error()))
		return nil
	}
	if msg.EventType != "play.completed" {
		p.log.InfoContext(ctx, "skipping non-completed play event", slog.String("event_type", msg.EventType))
		return nil
	}
	e := &domain.PlayEvent{
		UserID:     msg.UserID,
		TrackID:    msg.TrackID,
		SessionID:  msg.SessionID,
		EventType:  msg.EventType,
		OccurredAt: time.Unix(msg.OccurredAt, 0).UTC(),
	}
	return p.repo.Save(ctx, e)
}

func (p *PlayEventsConsumer) Close() error {
	return p.c.Close()
}
