package port

import (
	"context"

	"github.com/faqears/faqears/services/recommendation-service/internal/domain"
)

type PlayEventRepository interface {
	Save(ctx context.Context, e *domain.PlayEvent) error
	TopPlayedByUser(ctx context.Context, userID string, limit int, days int) ([]string, error)
	ActiveTrackIDs(ctx context.Context, days int) ([]string, error)
	TopGlobal(ctx context.Context, limit int) ([]*domain.TrackRef, error)
	PlayedByUser(ctx context.Context, userID string, days int) (map[string]struct{}, error)
}

type SimilarityRepository interface {
	CoListeners(ctx context.Context, trackID string, limit int) ([]*domain.Similarity, error)
}

type Cache interface {
	GetSimilar(ctx context.Context, trackID string) ([]*domain.TrackRef, error)
	SetSimilar(ctx context.Context, trackID string, refs []*domain.TrackRef) error
	GetDaily(ctx context.Context, userID string) ([]*domain.TrackRef, error)
	SetDaily(ctx context.Context, userID string, refs []*domain.TrackRef) error
	GetDiscoverWeekly(ctx context.Context, userID string) ([]*domain.TrackRef, error)
	SetDiscoverWeekly(ctx context.Context, userID string, refs []*domain.TrackRef) error
	GetTopGlobal(ctx context.Context) ([]*domain.TrackRef, error)
	SetTopGlobal(ctx context.Context, refs []*domain.TrackRef) error
}

type ConsumerHandler interface {
	Handle(ctx context.Context, key, value []byte, headers map[string]string) error
}
