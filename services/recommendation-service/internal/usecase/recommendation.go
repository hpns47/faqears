package usecase

import (
	"context"
	"log/slog"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/recommendation-service/internal/domain"
	"github.com/faqears/faqears/services/recommendation-service/internal/port"
)

const (
	defaultLimit      = 20
	seedTracksPerUser = 5
	simPerSeed        = 10
	topFallbackLimit  = 20
)

type Recommendation struct {
	plays  port.PlayEventRepository
	sim    port.SimilarityRepository
	cache  port.Cache
	log    *slog.Logger
}

func NewRecommendation(
	plays port.PlayEventRepository,
	sim port.SimilarityRepository,
	cache port.Cache,
	log *slog.Logger,
) *Recommendation {
	return &Recommendation{
		plays: plays,
		sim:   sim,
		cache: cache,
		log:   log,
	}
}

func (r *Recommendation) GetDailyMix(ctx context.Context, userID string, limit int32) ([]*domain.TrackRef, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if cached, err := r.cache.GetDaily(ctx, userID); err == nil && len(cached) > 0 {
		return cached, nil
	}

	seeds, err := r.plays.TopPlayedByUser(ctx, userID, seedTracksPerUser, 14)
	if err != nil {
		return nil, errs.Internal("top played by user", err)
	}
	if len(seeds) == 0 {
		return r.GetTopTracks(ctx, limit)
	}

	played, err := r.plays.PlayedByUser(ctx, userID, 14)
	if err != nil {
		return nil, errs.Internal("played by user", err)
	}

	result := r.buildMix(ctx, seeds, played, int(limit))
	if len(result) == 0 {
		return r.GetTopTracks(ctx, limit)
	}

	if err := r.cache.SetDaily(ctx, userID, result); err != nil {
		r.log.WarnContext(ctx, "cache set daily failed", slog.String("user_id", userID), slog.String("error", err.Error()))
	}
	return result, nil
}

func (r *Recommendation) GetDiscoverWeekly(ctx context.Context, userID string, limit int32) ([]*domain.TrackRef, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if cached, err := r.cache.GetDiscoverWeekly(ctx, userID); err == nil && len(cached) > 0 {
		return cached, nil
	}

	seeds, err := r.plays.TopPlayedByUser(ctx, userID, seedTracksPerUser, 30)
	if err != nil {
		return nil, errs.Internal("top played by user (weekly)", err)
	}
	if len(seeds) == 0 {
		return r.GetTopTracks(ctx, limit)
	}

	played, err := r.plays.PlayedByUser(ctx, userID, 30)
	if err != nil {
		return nil, errs.Internal("played by user (weekly)", err)
	}

	result := r.buildMix(ctx, seeds, played, int(limit))
	if len(result) == 0 {
		return r.GetTopTracks(ctx, limit)
	}

	if err := r.cache.SetDiscoverWeekly(ctx, userID, result); err != nil {
		r.log.WarnContext(ctx, "cache set discover_weekly failed", slog.String("user_id", userID), slog.String("error", err.Error()))
	}
	return result, nil
}

func (r *Recommendation) GetSimilarTracks(ctx context.Context, trackID string, limit int32) ([]*domain.TrackRef, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if cached, err := r.cache.GetSimilar(ctx, trackID); err == nil && len(cached) > 0 {
		return cached, nil
	}

	sims, err := r.sim.CoListeners(ctx, trackID, int(limit))
	if err != nil {
		return nil, errs.Internal("co listeners", err)
	}

	refs := make([]*domain.TrackRef, len(sims))
	for i, s := range sims {
		refs[i] = &domain.TrackRef{TrackID: s.TrackID, Score: float64(s.CoCount)}
	}

	if err := r.cache.SetSimilar(ctx, trackID, refs); err != nil {
		r.log.WarnContext(ctx, "cache set similar failed", slog.String("track_id", trackID), slog.String("error", err.Error()))
	}
	return refs, nil
}

func (r *Recommendation) GetTopTracks(ctx context.Context, limit int32) ([]*domain.TrackRef, error) {
	if limit <= 0 {
		limit = defaultLimit
	}
	if cached, err := r.cache.GetTopGlobal(ctx); err == nil && len(cached) > 0 {
		return cached, nil
	}

	refs, err := r.plays.TopGlobal(ctx, int(limit))
	if err != nil {
		return nil, errs.Internal("top global", err)
	}

	if err := r.cache.SetTopGlobal(ctx, refs); err != nil {
		r.log.WarnContext(ctx, "cache set top_global failed", slog.String("error", err.Error()))
	}
	return refs, nil
}

func (r *Recommendation) buildMix(ctx context.Context, seeds []string, played map[string]struct{}, limit int) []*domain.TrackRef {
	seen := make(map[string]struct{})
	for k := range played {
		seen[k] = struct{}{}
	}
	var result []*domain.TrackRef
	for _, seed := range seeds {
		cached, err := r.cache.GetSimilar(ctx, seed)
		if err != nil || len(cached) == 0 {
			continue
		}
		for _, ref := range cached {
			if _, ok := seen[ref.TrackID]; ok {
				continue
			}
			seen[ref.TrackID] = struct{}{}
			result = append(result, ref)
			if len(result) >= limit {
				return result
			}
			if len(result) >= simPerSeed {
				break
			}
		}
	}
	return result
}
