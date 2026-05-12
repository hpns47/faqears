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

func (p *Publisher) Close() error {
	return p.p.Close()
}

type playlistCreatedEvent struct {
	PlaylistID string `json:"playlist_id"`
	OwnerID    string `json:"owner_id"`
	OccurredAt int64  `json:"occurred_at"`
}

type trackAddedEvent struct {
	PlaylistID string `json:"playlist_id"`
	TrackID    string `json:"track_id"`
	AddedBy    string `json:"added_by"`
	OccurredAt int64  `json:"occurred_at"`
}

func (p *Publisher) PlaylistCreated(ctx context.Context, playlistID, ownerID string) error {
	evt := playlistCreatedEvent{
		PlaylistID: playlistID,
		OwnerID:    ownerID,
		OccurredAt: time.Now().UTC().Unix(),
	}
	value, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": "playlist.created"}
	return p.p.Publish(ctx, []byte(playlistID), value, headers)
}

func (p *Publisher) TrackAdded(ctx context.Context, playlistID, trackID, addedBy string) error {
	evt := trackAddedEvent{
		PlaylistID: playlistID,
		TrackID:    trackID,
		AddedBy:    addedBy,
		OccurredAt: time.Now().UTC().Unix(),
	}
	value, err := json.Marshal(evt)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": "playlist.track_added"}
	return p.p.Publish(ctx, []byte(playlistID), value, headers)
}
