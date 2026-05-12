package usecase

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"log/slog"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/services/playlist-service/internal/domain"
	"github.com/faqears/faqears/services/playlist-service/internal/port"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

type Playlist struct {
	playlists port.PlaylistRepository
	tracks    port.TrackRepository
	cache     port.Cache
	events    port.EventPublisher
	cacheTTL  time.Duration
}

func NewPlaylist(
	playlists port.PlaylistRepository,
	tracks port.TrackRepository,
	cache port.Cache,
	events port.EventPublisher,
	cacheTTL time.Duration,
) *Playlist {
	return &Playlist{
		playlists: playlists,
		tracks:    tracks,
		cache:     cache,
		events:    events,
		cacheTTL:  cacheTTL,
	}
}

func generatePermalink() (string, error) {
	b := make([]byte, 10)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(base62Chars))))
		if err != nil {
			return "", err
		}
		b[i] = base62Chars[n.Int64()]
	}
	return string(b), nil
}

func callerID(ctx context.Context, protoCallerID string) (string, error) {
	if protoCallerID != "" {
		return protoCallerID, nil
	}
	id := grpcx.UserIDFromCtx(ctx)
	if id == "" {
		return "", errs.Unauthenticated("caller identity required")
	}
	return id, nil
}

func (u *Playlist) canWrite(ctx context.Context, p *domain.Playlist, callerID string) error {
	if p.OwnerID == callerID {
		return nil
	}
	ok, err := u.playlists.IsCollaborator(ctx, p.ID, callerID)
	if err != nil {
		return errs.Internal("check collaborator", err)
	}
	if !ok {
		return domain.ErrPermissionDenied
	}
	return nil
}

func (u *Playlist) enrichPlaylist(ctx context.Context, p *domain.Playlist) error {
	colls, err := u.playlists.GetCollaboratorIDs(ctx, p.ID)
	if err != nil {
		return errs.Internal("get collaborators", err)
	}
	p.CollaboratorIDs = colls

	tracks, err := u.tracks.ListTracks(ctx, p.ID)
	if err != nil {
		return errs.Internal("list tracks", err)
	}
	p.Tracks = tracks
	return nil
}

func (u *Playlist) getFromCache(ctx context.Context, id string) (*domain.Playlist, bool) {
	p, err := u.cache.GetPlaylist(ctx, id)
	if err != nil || p == nil {
		return nil, false
	}
	return p, true
}

func (u *Playlist) setCache(ctx context.Context, p *domain.Playlist) {
	if err := u.cache.SetPlaylist(ctx, p, u.cacheTTL); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "cache set playlist failed",
			slog.String("playlist_id", p.ID),
			slog.String("error", err.Error()),
		)
	}
	if err := u.cache.SetPlaylistByPermalink(ctx, p, u.cacheTTL); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "cache set playlist permalink failed",
			slog.String("permalink", p.Permalink),
			slog.String("error", err.Error()),
		)
	}
}

func (u *Playlist) invalidateCache(ctx context.Context, p *domain.Playlist) {
	if err := u.cache.InvalidatePlaylist(ctx, p.ID, p.Permalink); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "cache invalidate failed",
			slog.String("playlist_id", p.ID),
			slog.String("error", err.Error()),
		)
	}
}

func (u *Playlist) CreatePlaylist(ctx context.Context, ownerID, name, description string, public bool) (*domain.Playlist, error) {
	if ownerID == "" {
		ownerID = grpcx.UserIDFromCtx(ctx)
	}
	if ownerID == "" {
		return nil, errs.Unauthenticated("owner identity required")
	}
	permalink, err := generatePermalink()
	if err != nil {
		return nil, errs.Internal("generate permalink", err)
	}
	now := time.Now().UTC()
	p := &domain.Playlist{
		ID:          uuid.NewString(),
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
		Public:      public,
		Permalink:   permalink,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := u.playlists.Create(ctx, p); err != nil {
		return nil, errs.Internal("create playlist", err)
	}
	p.CollaboratorIDs = []string{}
	p.Tracks = []*domain.PlaylistTrack{}
	u.setCache(ctx, p)
	if err := u.events.PlaylistCreated(ctx, p.ID, ownerID); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "publish playlist.created failed",
			slog.String("playlist_id", p.ID),
			slog.String("error", err.Error()),
		)
	}
	return p, nil
}

func (u *Playlist) GetPlaylist(ctx context.Context, playlistID string) (*domain.Playlist, error) {
	if p, ok := u.getFromCache(ctx, playlistID); ok {
		return p, nil
	}
	p, err := u.playlists.GetByID(ctx, playlistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errs.Internal("get playlist", err)
	}
	if err := u.enrichPlaylist(ctx, p); err != nil {
		return nil, err
	}
	u.setCache(ctx, p)
	return p, nil
}

func (u *Playlist) GetPlaylistByPermalink(ctx context.Context, permalink string) (*domain.Playlist, error) {
	cached, err := u.cache.GetPlaylistByPermalink(ctx, permalink)
	if err == nil && cached != nil {
		return cached, nil
	}
	p, err := u.playlists.GetByPermalink(ctx, permalink)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errs.Internal("get playlist by permalink", err)
	}
	if err := u.enrichPlaylist(ctx, p); err != nil {
		return nil, err
	}
	u.setCache(ctx, p)
	return p, nil
}

func (u *Playlist) UpdatePlaylist(ctx context.Context, playlistID, callerIDStr, name, description string, public bool) (*domain.Playlist, error) {
	caller, err := callerID(ctx, callerIDStr)
	if err != nil {
		return nil, err
	}
	p, err := u.playlists.GetByID(ctx, playlistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errs.Internal("get playlist", err)
	}
	if err := u.canWrite(ctx, p, caller); err != nil {
		return nil, err
	}
	p.Name = name
	p.Description = description
	p.Public = public
	p.UpdatedAt = time.Now().UTC()
	if err := u.playlists.Update(ctx, p); err != nil {
		return nil, errs.Internal("update playlist", err)
	}
	u.invalidateCache(ctx, p)
	if err := u.enrichPlaylist(ctx, p); err != nil {
		return nil, err
	}
	u.setCache(ctx, p)
	return p, nil
}

func (u *Playlist) DeletePlaylist(ctx context.Context, playlistID, callerIDStr string) error {
	caller, err := callerID(ctx, callerIDStr)
	if err != nil {
		return err
	}
	p, err := u.playlists.GetByID(ctx, playlistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return errs.Internal("get playlist", err)
	}
	if p.OwnerID != caller {
		return domain.ErrPermissionDenied
	}
	u.invalidateCache(ctx, p)
	if err := u.playlists.Delete(ctx, playlistID); err != nil {
		return errs.Internal("delete playlist", err)
	}
	return nil
}

func midpointRank(a, b string) string {
	fa, _ := new(big.Float).SetString(a)
	fb, _ := new(big.Float).SetString(b)
	sum := new(big.Float).Add(fa, fb)
	mid := new(big.Float).Quo(sum, big.NewFloat(2))
	return mid.Text('f', 20)
}

func (u *Playlist) AddTrack(ctx context.Context, playlistID, trackID, afterTrackID, callerIDStr string) (*domain.Playlist, error) {
	caller, err := callerID(ctx, callerIDStr)
	if err != nil {
		return nil, err
	}
	p, err := u.playlists.GetByID(ctx, playlistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errs.Internal("get playlist", err)
	}
	if err := u.canWrite(ctx, p, caller); err != nil {
		return nil, err
	}
	existing, err := u.tracks.GetTrack(ctx, playlistID, trackID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, errs.Internal("check track existence", err)
	}
	if existing != nil {
		return nil, domain.ErrTrackAlreadyInPlaylist
	}

	var rank string
	if afterTrackID == "" {
		maxRank, err := u.tracks.GetMaxRank(ctx, playlistID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.Internal("get max rank", err)
		}
		if maxRank == "" {
			rank = "1.00000000000000000000"
		} else {
			fr, _ := new(big.Float).SetString(maxRank)
			rank = new(big.Float).Add(fr, big.NewFloat(1.0)).Text('f', 20)
		}
	} else {
		afterTrack, err := u.tracks.GetTrack(ctx, playlistID, afterTrackID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrTrackNotInPlaylist
			}
			return nil, errs.Internal("get after track", err)
		}
		nextRank, hasNext, err := u.tracks.GetRankAfterTrack(ctx, playlistID, afterTrackID)
		if err != nil {
			return nil, errs.Internal("get next rank", err)
		}
		if hasNext {
			rank = midpointRank(afterTrack.Rank, nextRank)
		} else {
			fr, _ := new(big.Float).SetString(afterTrack.Rank)
			rank = new(big.Float).Add(fr, big.NewFloat(1.0)).Text('f', 20)
		}
	}

	pt := &domain.PlaylistTrack{
		PlaylistID: playlistID,
		TrackID:    trackID,
		Rank:       rank,
		AddedBy:    caller,
		AddedAt:    time.Now().UTC(),
	}
	if err := u.tracks.AddTrack(ctx, pt); err != nil {
		return nil, errs.Internal("add track", err)
	}
	if err := u.cache.AddTrackPlaylistRef(ctx, trackID, playlistID); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "add track playlist ref failed",
			slog.String("track_id", trackID),
			slog.String("error", err.Error()),
		)
	}
	u.invalidateCache(ctx, p)
	if err := u.enrichPlaylist(ctx, p); err != nil {
		return nil, err
	}
	u.setCache(ctx, p)
	if err := u.events.TrackAdded(ctx, playlistID, trackID, caller); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "publish track_added failed",
			slog.String("playlist_id", playlistID),
			slog.String("error", err.Error()),
		)
	}
	return p, nil
}

func (u *Playlist) RemoveTrack(ctx context.Context, playlistID, trackID, callerIDStr string) (*domain.Playlist, error) {
	caller, err := callerID(ctx, callerIDStr)
	if err != nil {
		return nil, err
	}
	p, err := u.playlists.GetByID(ctx, playlistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errs.Internal("get playlist", err)
	}
	if err := u.canWrite(ctx, p, caller); err != nil {
		return nil, err
	}
	existing, err := u.tracks.GetTrack(ctx, playlistID, trackID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTrackNotInPlaylist
		}
		return nil, errs.Internal("get track", err)
	}
	if existing == nil {
		return nil, domain.ErrTrackNotInPlaylist
	}
	if err := u.tracks.RemoveTrack(ctx, playlistID, trackID); err != nil {
		return nil, errs.Internal("remove track", err)
	}
	if err := u.cache.RemoveTrackPlaylistRef(ctx, trackID, playlistID); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "remove track playlist ref failed",
			slog.String("track_id", trackID),
			slog.String("error", err.Error()),
		)
	}
	u.invalidateCache(ctx, p)
	if err := u.enrichPlaylist(ctx, p); err != nil {
		return nil, err
	}
	u.setCache(ctx, p)
	return p, nil
}

func (u *Playlist) ReorderTrack(ctx context.Context, playlistID, trackID, afterTrackID, callerIDStr string) (*domain.Playlist, error) {
	caller, err := callerID(ctx, callerIDStr)
	if err != nil {
		return nil, err
	}
	p, err := u.playlists.GetByID(ctx, playlistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, errs.Internal("get playlist", err)
	}
	if err := u.canWrite(ctx, p, caller); err != nil {
		return nil, err
	}
	track, err := u.tracks.GetTrack(ctx, playlistID, trackID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTrackNotInPlaylist
		}
		return nil, errs.Internal("get track", err)
	}
	if track == nil {
		return nil, domain.ErrTrackNotInPlaylist
	}

	var newRank string
	if afterTrackID == "" {
		minRank, err := u.tracks.GetMinRank(ctx, playlistID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.Internal("get min rank", err)
		}
		if minRank == "" {
			newRank = "1.00000000000000000000"
		} else {
			fr, _ := new(big.Float).SetString(minRank)
			newRank = new(big.Float).Sub(fr, big.NewFloat(1.0)).Text('f', 20)
		}
	} else {
		afterTrack, err := u.tracks.GetTrack(ctx, playlistID, afterTrackID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, domain.ErrTrackNotInPlaylist
			}
			return nil, errs.Internal("get after track", err)
		}
		nextRank, hasNext, err := u.tracks.GetRankAfterTrack(ctx, playlistID, afterTrackID)
		if err != nil {
			return nil, errs.Internal("get next rank", err)
		}
		if hasNext {
			newRank = midpointRank(afterTrack.Rank, nextRank)
		} else {
			fr, _ := new(big.Float).SetString(afterTrack.Rank)
			newRank = new(big.Float).Add(fr, big.NewFloat(1.0)).Text('f', 20)
		}
	}

	if err := u.tracks.UpdateRank(ctx, playlistID, trackID, newRank); err != nil {
		return nil, errs.Internal("update track rank", err)
	}
	u.invalidateCache(ctx, p)
	if err := u.enrichPlaylist(ctx, p); err != nil {
		return nil, err
	}
	u.setCache(ctx, p)
	return p, nil
}

func (u *Playlist) ListUserPlaylists(ctx context.Context, userID string, limit, offset int32, includeCollaborations bool) ([]*domain.Playlist, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	owned, err := u.playlists.ListByOwner(ctx, userID, limit, offset)
	if err != nil {
		return nil, errs.Internal("list playlists by owner", err)
	}
	result := owned
	if includeCollaborations {
		collab, err := u.playlists.ListByCollaborator(ctx, userID, limit, offset)
		if err != nil {
			return nil, errs.Internal("list playlists by collaborator", err)
		}
		result = append(result, collab...)
	}
	for _, p := range result {
		if err := u.enrichPlaylist(ctx, p); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (u *Playlist) AddCollaborator(ctx context.Context, playlistID, collaboratorID, callerIDStr string) error {
	caller, err := callerID(ctx, callerIDStr)
	if err != nil {
		return err
	}
	p, err := u.playlists.GetByID(ctx, playlistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return errs.Internal("get playlist", err)
	}
	if p.OwnerID != caller {
		return domain.ErrPermissionDenied
	}
	c := &domain.Collaborator{
		PlaylistID: playlistID,
		UserID:     collaboratorID,
		AddedAt:    time.Now().UTC(),
	}
	if err := u.playlists.AddCollaborator(ctx, c); err != nil {
		return errs.Internal("add collaborator", err)
	}
	u.invalidateCache(ctx, p)
	return nil
}

func (u *Playlist) RemoveCollaborator(ctx context.Context, playlistID, collaboratorID, callerIDStr string) error {
	caller, err := callerID(ctx, callerIDStr)
	if err != nil {
		return err
	}
	p, err := u.playlists.GetByID(ctx, playlistID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return errs.Internal("get playlist", err)
	}
	if p.OwnerID != caller {
		return domain.ErrPermissionDenied
	}
	if err := u.playlists.RemoveCollaborator(ctx, playlistID, collaboratorID); err != nil {
		return errs.Internal("remove collaborator", err)
	}
	u.invalidateCache(ctx, p)
	return nil
}

func (u *Playlist) HandleTrackUpdated(ctx context.Context, trackID string) error {
	playlistIDs, err := u.cache.GetPlaylistsForTrack(ctx, trackID)
	if err != nil || len(playlistIDs) == 0 {
		dbIDs, dbErr := u.tracks.ListPlaylistsForTrack(ctx, trackID)
		if dbErr != nil {
			return errs.Internal("list playlists for track", dbErr)
		}
		playlistIDs = dbIDs
	}
	for _, pid := range playlistIDs {
		p, dbErr := u.playlists.GetByID(ctx, pid)
		if dbErr != nil {
			continue
		}
		u.invalidateCache(ctx, p)
	}
	return nil
}

var _ = json.Marshal
