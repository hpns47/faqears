package handler

import (
	"encoding/json"
	"io"
	"net/http"

	catalogv1 "github.com/faqears/faqears/gen/go/catalog/v1"
	streamingv1 "github.com/faqears/faqears/gen/go/streaming/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/go-chi/chi/v5"
)

type StreamingHandler struct {
	c       streamingv1.StreamingServiceClient
	catalog catalogv1.CatalogServiceClient
}

func NewStreamingHandler(c streamingv1.StreamingServiceClient, catalog catalogv1.CatalogServiceClient) *StreamingHandler {
	return &StreamingHandler{c: c, catalog: catalog}
}

func (h *StreamingHandler) GetStreamURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.GetStreamURL(ctx, &streamingv1.GetStreamURLRequest{TrackId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *StreamingHandler) RedirectToStream(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.GetStreamURL(ctx, &streamingv1.GetStreamURLRequest{TrackId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	http.Redirect(w, r, resp.StreamUrl, http.StatusFound)
}

func (h *StreamingHandler) RegisterPlay(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TrackID string `json:"track_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.RegisterPlay(ctx, &streamingv1.RegisterPlayRequest{TrackId: body.TrackID})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"session_id": resp.SessionId})
}

func (h *StreamingHandler) RegisterSkip(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TrackID     string `json:"track_id"`
		SessionID   string `json:"session_id"`
		PositionSec int32  `json:"position_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	_, err := h.c.RegisterSkip(ctx, &streamingv1.RegisterSkipRequest{
		TrackId:     body.TrackID,
		SessionId:   body.SessionID,
		PositionSec: body.PositionSec,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (h *StreamingHandler) RegisterComplete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		TrackID     string `json:"track_id"`
		SessionID   string `json:"session_id"`
		DurationSec int32  `json:"duration_sec"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	_, err := h.c.RegisterComplete(ctx, &streamingv1.RegisterCompleteRequest{
		TrackId:     body.TrackID,
		SessionId:   body.SessionID,
		DurationSec: body.DurationSec,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (h *StreamingHandler) UploadAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.UploadAudio(ctx, &streamingv1.UploadAudioRequest{
		TrackId:     id,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *StreamingHandler) UploadUserAudio(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	identity := callerIdentity(r.Context())
	if identity.UserID == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	trackResp, err := h.catalog.GetTrack(r.Context(), &catalogv1.GetTrackRequest{TrackId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	if trackResp.Track == nil {
		writeError(w, http.StatusNotFound, "track not found")
		return
	}
	if trackResp.Track.OwnerUserId != identity.UserID {
		writeError(w, http.StatusForbidden, "you don't own this track")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 50<<20)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read file")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.UploadAudio(ctx, &streamingv1.UploadAudioRequest{
		TrackId:     id,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}
