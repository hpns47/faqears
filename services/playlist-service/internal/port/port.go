package port

import (
	"context"
	"time"

	"github.com/faqears/faqears/services/playlist-service/internal/domain"
)

type PlaylistRepository interface {
	Create(ctx context.Context, p *domain.Playlist) error
	GetByID(ctx context.Context, id string) (*domain.Playlist, error)
	GetByPermalink(ctx context.Context, permalink string) (*domain.Playlist, error)
	Update(ctx context.Context, p *domain.Playlist) error
	Delete(ctx context.Context, id string) error
	ListByOwner(ctx context.Context, ownerID string, limit, offset int32) ([]*domain.Playlist, error)
	ListByCollaborator(ctx context.Context, userID string, limit, offset int32) ([]*domain.Playlist, error)
	IsCollaborator(ctx context.Context, playlistID, userID string) (bool, error)
	AddCollaborator(ctx context.Context, c *domain.Collaborator) error
	RemoveCollaborator(ctx context.Context, playlistID, userID string) error
	GetCollaboratorIDs(ctx context.Context, playlistID string) ([]string, error)
}

type TrackRepository interface {
	AddTrack(ctx context.Context, t *domain.PlaylistTrack) error
	RemoveTrack(ctx context.Context, playlistID, trackID string) error
	GetTrack(ctx context.Context, playlistID, trackID string) (*domain.PlaylistTrack, error)
	ListTracks(ctx context.Context, playlistID string) ([]*domain.PlaylistTrack, error)
	GetMaxRank(ctx context.Context, playlistID string) (string, error)
	GetMinRank(ctx context.Context, playlistID string) (string, error)
	GetRankAfterTrack(ctx context.Context, playlistID, trackID string) (string, bool, error)
	UpdateRank(ctx context.Context, playlistID, trackID, rank string) error
	ListPlaylistsForTrack(ctx context.Context, trackID string) ([]string, error)
}

type Cache interface {
	GetPlaylist(ctx context.Context, id string) (*domain.Playlist, error)
	SetPlaylist(ctx context.Context, p *domain.Playlist, ttl time.Duration) error
	GetPlaylistByPermalink(ctx context.Context, permalink string) (*domain.Playlist, error)
	SetPlaylistByPermalink(ctx context.Context, p *domain.Playlist, ttl time.Duration) error
	InvalidatePlaylist(ctx context.Context, id, permalink string) error
	AddTrackPlaylistRef(ctx context.Context, trackID, playlistID string) error
	RemoveTrackPlaylistRef(ctx context.Context, trackID, playlistID string) error
	GetPlaylistsForTrack(ctx context.Context, trackID string) ([]string, error)
}

type EventPublisher interface {
	PlaylistCreated(ctx context.Context, playlistID, ownerID string) error
	TrackAdded(ctx context.Context, playlistID, trackID, addedBy string) error
	Close() error
}
