package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/services/catalog-service/internal/domain"
	"github.com/faqears/faqears/services/catalog-service/internal/port"
)

type Catalog struct {
	artists  port.ArtistRepository
	albums   port.AlbumRepository
	tracks   port.TrackRepository
	cache    port.Cache
	events   port.EventPublisher
	cacheTTL time.Duration
}

func NewCatalog(
	artists port.ArtistRepository,
	albums port.AlbumRepository,
	tracks port.TrackRepository,
	cache port.Cache,
	events port.EventPublisher,
	cacheTTL time.Duration,
) *Catalog {
	return &Catalog{
		artists:  artists,
		albums:   albums,
		tracks:   tracks,
		cache:    cache,
		events:   events,
		cacheTTL: cacheTTL,
	}
}

func (c *Catalog) GetArtist(ctx context.Context, artistID string) (*domain.Artist, error) {
	if a, err := c.cache.GetArtist(ctx, artistID); err == nil && a != nil {
		return a, nil
	}
	a, err := c.artists.GetByID(ctx, artistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrArtistNotFound
		}
		return nil, errs.Internal("get artist", err)
	}
	if err := c.cache.SetArtist(ctx, a, c.cacheTTL); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "cache set artist failed",
			slog.String("artist_id", artistID),
			slog.String("error", err.Error()),
		)
	}
	return a, nil
}

func (c *Catalog) GetAlbum(ctx context.Context, albumID string) (*domain.Album, error) {
	if a, err := c.cache.GetAlbum(ctx, albumID); err == nil && a != nil {
		return a, nil
	}
	a, err := c.albums.GetByID(ctx, albumID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrAlbumNotFound
		}
		return nil, errs.Internal("get album", err)
	}
	if err := c.cache.SetAlbum(ctx, a, c.cacheTTL); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "cache set album failed",
			slog.String("album_id", albumID),
			slog.String("error", err.Error()),
		)
	}
	return a, nil
}

func (c *Catalog) GetTrack(ctx context.Context, trackID string) (*domain.Track, error) {
	if t, err := c.cache.GetTrack(ctx, trackID); err == nil && t != nil {
		return t, nil
	}
	t, err := c.tracks.GetByID(ctx, trackID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTrackNotFound
		}
		return nil, errs.Internal("get track", err)
	}
	if err := c.cache.SetTrack(ctx, t, c.cacheTTL); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "cache set track failed",
			slog.String("track_id", trackID),
			slog.String("error", err.Error()),
		)
	}
	return t, nil
}

func (c *Catalog) ListTracksByAlbum(ctx context.Context, albumID string, limit, offset int32) ([]*domain.Track, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	tracks, err := c.tracks.ListByAlbum(ctx, albumID, limit, offset)
	if err != nil {
		return nil, errs.Internal("list tracks by album", err)
	}
	return tracks, nil
}

func (c *Catalog) ListAlbumsByArtist(ctx context.Context, artistID string, limit, offset int32) ([]*domain.Album, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	albums, err := c.albums.ListByArtist(ctx, artistID, limit, offset)
	if err != nil {
		return nil, errs.Internal("list albums by artist", err)
	}
	return albums, nil
}

func (c *Catalog) Search(ctx context.Context, query string, limit int32) ([]*domain.Track, []*domain.Album, []*domain.Artist, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	tracks, err := c.tracks.Search(ctx, query, limit)
	if err != nil {
		return nil, nil, nil, errs.Internal("search tracks", err)
	}
	albums, err := c.albums.Search(ctx, query, limit)
	if err != nil {
		return nil, nil, nil, errs.Internal("search albums", err)
	}
	artists, err := c.artists.Search(ctx, query, limit)
	if err != nil {
		return nil, nil, nil, errs.Internal("search artists", err)
	}
	return tracks, albums, artists, nil
}

func (c *Catalog) IngestArtist(ctx context.Context, name, country, biography string) (string, error) {
	a := &domain.Artist{
		ID:        uuid.NewString(),
		Name:      name,
		Country:   country,
		Biography: biography,
		CreatedAt: time.Now().UTC(),
	}
	if err := c.artists.Create(ctx, a); err != nil {
		return "", errs.Internal("create artist", err)
	}
	return a.ID, nil
}

func (c *Catalog) IngestAlbum(ctx context.Context, artistID, title string, year int32, coverURL string) (string, error) {
	a := &domain.Album{
		ID:        uuid.NewString(),
		ArtistID:  artistID,
		Title:     title,
		Year:      year,
		CoverURL:  coverURL,
		CreatedAt: time.Now().UTC(),
	}
	if err := c.albums.Create(ctx, a); err != nil {
		return "", errs.Internal("create album", err)
	}
	if err := c.events.AlbumAdded(ctx, a.ID, artistID, title); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "publish album_added failed",
			slog.String("album_id", a.ID),
			slog.String("error", err.Error()),
		)
	}
	return a.ID, nil
}

func (c *Catalog) IngestTrack(ctx context.Context, albumID, artistID, title string, durationSec int32, isrc string, genres []string, userGenerated bool, ownerUserID string) (string, error) {
	t := &domain.Track{
		ID:            uuid.NewString(),
		AlbumID:       albumID,
		ArtistID:      artistID,
		Title:         title,
		DurationSec:   durationSec,
		ISRC:          isrc,
		Genres:        genres,
		UserGenerated: userGenerated,
		OwnerUserID:   ownerUserID,
		CreatedAt:     time.Now().UTC(),
	}
	if err := c.tracks.Create(ctx, t); err != nil {
		return "", errs.Internal("create track", err)
	}
	if err := c.events.TrackAdded(ctx, t.ID, albumID, artistID, title); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "publish track_added failed",
			slog.String("track_id", t.ID),
			slog.String("error", err.Error()),
		)
	}
	return t.ID, nil
}
