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

type playEvent struct {
	EventType  string `json:"event_type"`
	UserID     string `json:"user_id"`
	TrackID    string `json:"track_id"`
	SessionID  string `json:"session_id"`
	OccurredAt int64  `json:"occurred_at"`
}

type skipEvent struct {
	EventType   string `json:"event_type"`
	UserID      string `json:"user_id"`
	TrackID     string `json:"track_id"`
	SessionID   string `json:"session_id"`
	PositionSec int32  `json:"position_sec"`
	OccurredAt  int64  `json:"occurred_at"`
}

type completeEvent struct {
	EventType   string `json:"event_type"`
	UserID      string `json:"user_id"`
	TrackID     string `json:"track_id"`
	SessionID   string `json:"session_id"`
	DurationSec int32  `json:"duration_sec"`
	OccurredAt  int64  `json:"occurred_at"`
}

func (p *Publisher) PlayStarted(ctx context.Context, userID, trackID, sessionID string) error {
	msg := playEvent{
		EventType:  "play.started",
		UserID:     userID,
		TrackID:    trackID,
		SessionID:  sessionID,
		OccurredAt: time.Now().UTC().Unix(),
	}
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": "play.started"}
	return p.p.Publish(ctx, []byte(sessionID), value, headers)
}

func (p *Publisher) PlaySkipped(ctx context.Context, userID, trackID, sessionID string, positionSec int32) error {
	msg := skipEvent{
		EventType:   "play.skipped",
		UserID:      userID,
		TrackID:     trackID,
		SessionID:   sessionID,
		PositionSec: positionSec,
		OccurredAt:  time.Now().UTC().Unix(),
	}
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": "play.skipped"}
	return p.p.Publish(ctx, []byte(sessionID), value, headers)
}

func (p *Publisher) PlayCompleted(ctx context.Context, userID, trackID, sessionID string, durationSec int32) error {
	msg := completeEvent{
		EventType:   "play.completed",
		UserID:      userID,
		TrackID:     trackID,
		SessionID:   sessionID,
		DurationSec: durationSec,
		OccurredAt:  time.Now().UTC().Unix(),
	}
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": "play.completed"}
	return p.p.Publish(ctx, []byte(sessionID), value, headers)
}

func (p *Publisher) Close() error {
	return p.p.Close()
}
