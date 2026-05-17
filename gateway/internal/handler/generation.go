package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	generationv1 "github.com/faqears/faqears/gen/go/generation/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/go-chi/chi/v5"
)

type GenerationHandler struct {
	c generationv1.GenerationServiceClient
}

func NewGenerationHandler(c generationv1.GenerationServiceClient) *GenerationHandler {
	return &GenerationHandler{c: c}
}

func (h *GenerationHandler) GenerateLyrics(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Prompt   string `json:"prompt"`
		Genre    string `json:"genre"`
		Mood     string `json:"mood"`
		Language string `json:"language"`
		Title    string `json:"title"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.GenerateLyrics(ctx, &generationv1.GenerateLyricsRequest{
		UserId:   identity.UserID,
		Prompt:   body.Prompt,
		Genre:    body.Genre,
		Mood:     body.Mood,
		Language: body.Language,
		Title:    body.Title,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"job":    resp.Job,
		"lyrics": resp.Lyrics,
	})
}

func (h *GenerationHandler) ReviseLyrics(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "job_id")
	var body struct {
		Instruction string `json:"instruction"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.ReviseLyrics(ctx, &generationv1.ReviseLyricsRequest{
		JobId:       jobID,
		UserId:      identity.UserID,
		Instruction: body.Instruction,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"job":    resp.Job,
		"lyrics": resp.Lyrics,
	})
}

func (h *GenerationHandler) GenerateMusic(w http.ResponseWriter, r *http.Request) {
	var body struct {
		JobID string `json:"job_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.GenerateMusic(ctx, &generationv1.GenerateMusicRequest{
		JobId:  body.JobID,
		UserId: identity.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{"job": resp.Job})
}

func (h *GenerationHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "job_id")
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.GetJob(ctx, &generationv1.GetJobRequest{JobId: jobID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"job":            resp.Job,
		"lyrics_history": resp.LyricsHistory,
	})
}

func (h *GenerationHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	identity := callerIdentity(r.Context())
	q := r.URL.Query()
	limit := int32(20)
	offset := int32(0)
	if v := q.Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	if v := q.Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = int32(n)
		}
	}
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.ListUserJobs(ctx, &generationv1.ListUserJobsRequest{
		UserId: identity.UserID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"jobs": resp.Jobs})
}

func (h *GenerationHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	jobID := chi.URLParam(r, "job_id")
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.CancelJob(ctx, &generationv1.CancelJobRequest{JobId: jobID, UserId: identity.UserID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"job": resp.Job})
}

func (h *GenerationHandler) JobEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, _ := w.(http.Flusher)
	jobID := chi.URLParam(r, "job_id")
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			resp, err := h.c.GetJob(ctx, &generationv1.GetJobRequest{JobId: jobID})
			if err != nil {
				return
			}
			data, _ := json.Marshal(resp.GetJob())
			fmt.Fprintf(w, "data: %s\n\n", data)
			if flusher != nil {
				flusher.Flush()
			}
			if isTerminal(resp.GetJob().GetStatus()) {
				return
			}
		}
	}
}

func isTerminal(status string) bool {
	return status == "completed" || status == "failed" || status == "cancelled"
}
