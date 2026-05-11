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

type trackAddedEvent struct {
	EventType string `json:"event_type"`
	TrackID   string `json:"track_id"`
	AlbumID   string `json:"album_id"`
	ArtistID  string `json:"artist_id"`
	Title     string `json:"title"`
	OccurredAt int64 `json:"occurred_at"`
}

type albumAddedEvent struct {
	EventType  string `json:"event_type"`
	AlbumID    string `json:"album_id"`
	ArtistID   string `json:"artist_id"`
	Title      string `json:"title"`
	OccurredAt int64  `json:"occurred_at"`
}

func (p *Publisher) TrackAdded(ctx context.Context, trackID, albumID, artistID, title string) error {
	msg := trackAddedEvent{
		EventType:  "catalog.track_added",
		TrackID:    trackID,
		AlbumID:    albumID,
		ArtistID:   artistID,
		Title:      title,
		OccurredAt: time.Now().UTC().Unix(),
	}
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": "catalog.track_added"}
	return p.p.Publish(ctx, []byte(trackID), value, headers)
}

func (p *Publisher) AlbumAdded(ctx context.Context, albumID, artistID, title string) error {
	msg := albumAddedEvent{
		EventType:  "catalog.album_added",
		AlbumID:    albumID,
		ArtistID:   artistID,
		Title:      title,
		OccurredAt: time.Now().UTC().Unix(),
	}
	value, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	headers := map[string]string{"event_type": "catalog.album_added"}
	return p.p.Publish(ctx, []byte(albumID), value, headers)
}

func (p *Publisher) Close() error {
	return p.p.Close()
}
