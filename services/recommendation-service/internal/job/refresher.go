package job

import (
	"context"
	"log/slog"
	"time"

	"github.com/faqears/faqears/services/recommendation-service/internal/domain"
	"github.com/faqears/faqears/services/recommendation-service/internal/port"
)

const (
	simLimit   = 50
	activeDays = 30
)

type Refresher struct {
	plays    port.PlayEventRepository
	sim      port.SimilarityRepository
	cache    port.Cache
	interval time.Duration
	log      *slog.Logger
}

func NewRefresher(
	plays port.PlayEventRepository,
	sim port.SimilarityRepository,
	cache port.Cache,
	interval time.Duration,
	log *slog.Logger,
) *Refresher {
	return &Refresher{
		plays:    plays,
		sim:      sim,
		cache:    cache,
		interval: interval,
		log:      log,
	}
}

func (r *Refresher) Run(ctx context.Context) {
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()
	r.refresh(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.refresh(ctx)
		}
	}
}

func (r *Refresher) refresh(ctx context.Context) {
	trackIDs, err := r.plays.ActiveTrackIDs(ctx, activeDays)
	if err != nil {
		r.log.ErrorContext(ctx, "refresher: active track ids", slog.String("error", err.Error()))
		return
	}
	r.log.InfoContext(ctx, "refresher: rebuilding similarity", slog.Int("tracks", len(trackIDs)))
	for _, trackID := range trackIDs {
		sims, err := r.sim.CoListeners(ctx, trackID, simLimit)
		if err != nil {
			r.log.WarnContext(ctx, "refresher: co listeners", slog.String("track_id", trackID), slog.String("error", err.Error()))
			continue
		}
		refs := make([]*domain.TrackRef, len(sims))
		for i, s := range sims {
			refs[i] = &domain.TrackRef{TrackID: s.TrackID, Score: float64(s.CoCount)}
		}
		if err := r.cache.SetSimilar(ctx, trackID, refs); err != nil {
			r.log.WarnContext(ctx, "refresher: cache set similar", slog.String("track_id", trackID), slog.String("error", err.Error()))
		}
	}
	r.log.InfoContext(ctx, "refresher: similarity rebuild complete", slog.Int("tracks", len(trackIDs)))
}
