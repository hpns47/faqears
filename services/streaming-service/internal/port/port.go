package port

import (
	"context"
	"time"

	"github.com/faqears/faqears/services/streaming-service/internal/domain"
)

type AudioRepository interface {
	Upsert(ctx context.Context, a *domain.TrackAudio) error
	GetByTrackID(ctx context.Context, trackID string) (*domain.TrackAudio, error)
	Exists(ctx context.Context, trackID string) (bool, int64, error)
}

type SessionStore interface {
	Save(ctx context.Context, s *domain.Session, ttl time.Duration) error
	Get(ctx context.Context, sessionID string) (*domain.Session, error)
	Delete(ctx context.Context, sessionID string) error
}

type Storage interface {
	Put(ctx context.Context, key, contentType string, data []byte) (int64, error)
	PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error)
}

type EventPublisher interface {
	PlayStarted(ctx context.Context, userID, trackID, sessionID string) error
	PlaySkipped(ctx context.Context, userID, trackID, sessionID string, positionSec int32) error
	PlayCompleted(ctx context.Context, userID, trackID, sessionID string, durationSec int32) error
	Close() error
}
