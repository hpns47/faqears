package handler

import (
	"net/http"
	"strconv"

	recommendationv1 "github.com/faqears/faqears/gen/go/recommendation/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/go-chi/chi/v5"
)

type RecommendationHandler struct {
	c recommendationv1.RecommendationServiceClient
}

func NewRecommendationHandler(c recommendationv1.RecommendationServiceClient) *RecommendationHandler {
	return &RecommendationHandler{c: c}
}

func parseLimit(r *http.Request, def int32) int32 {
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			return int32(n)
		}
	}
	return def
}

func (h *RecommendationHandler) DailyMix(w http.ResponseWriter, r *http.Request) {
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.GetDailyMix(ctx, &recommendationv1.GetDailyMixRequest{
		UserId: identity.UserID,
		Limit:  parseLimit(r, 20),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tracks":       resp.Tracks,
		"generated_at": resp.GeneratedAt,
	})
}

func (h *RecommendationHandler) DiscoverWeekly(w http.ResponseWriter, r *http.Request) {
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.GetDiscoverWeekly(ctx, &recommendationv1.GetDiscoverWeeklyRequest{
		UserId: identity.UserID,
		Limit:  parseLimit(r, 20),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tracks":       resp.Tracks,
		"generated_at": resp.GeneratedAt,
	})
}

func (h *RecommendationHandler) SimilarTracks(w http.ResponseWriter, r *http.Request) {
	trackID := chi.URLParam(r, "track_id")
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.GetSimilarTracks(ctx, &recommendationv1.GetSimilarTracksRequest{
		TrackId: trackID,
		Limit:   parseLimit(r, 10),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": resp.Tracks})
}

func (h *RecommendationHandler) TopTracks(w http.ResponseWriter, r *http.Request) {
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.GetTopTracks(ctx, &recommendationv1.GetTopTracksRequest{
		Limit: parseLimit(r, 10),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": resp.Tracks})
}
