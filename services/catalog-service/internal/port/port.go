package port

import (
	"context"
	"time"

	"github.com/faqears/faqears/services/catalog-service/internal/domain"
)

type ArtistRepository interface {
	Create(ctx context.Context, a *domain.Artist) error
	GetByID(ctx context.Context, id string) (*domain.Artist, error)
	Search(ctx context.Context, query string, limit int32) ([]*domain.Artist, error)
}

type AlbumRepository interface {
	Create(ctx context.Context, a *domain.Album) error
	GetByID(ctx context.Context, id string) (*domain.Album, error)
	ListByArtist(ctx context.Context, artistID string, limit, offset int32) ([]*domain.Album, error)
	Search(ctx context.Context, query string, limit int32) ([]*domain.Album, error)
}

type TrackRepository interface {
	Create(ctx context.Context, t *domain.Track) error
	GetByID(ctx context.Context, id string) (*domain.Track, error)
	ListByAlbum(ctx context.Context, albumID string, limit, offset int32) ([]*domain.Track, error)
	Search(ctx context.Context, query string, limit int32) ([]*domain.Track, error)
}

type Cache interface {
	GetArtist(ctx context.Context, id string) (*domain.Artist, error)
	SetArtist(ctx context.Context, a *domain.Artist, ttl time.Duration) error
	GetAlbum(ctx context.Context, id string) (*domain.Album, error)
	SetAlbum(ctx context.Context, a *domain.Album, ttl time.Duration) error
	GetTrack(ctx context.Context, id string) (*domain.Track, error)
	SetTrack(ctx context.Context, t *domain.Track, ttl time.Duration) error
}

type EventPublisher interface {
	TrackAdded(ctx context.Context, trackID, albumID, artistID, title string) error
	AlbumAdded(ctx context.Context, albumID, artistID, title string) error
}
