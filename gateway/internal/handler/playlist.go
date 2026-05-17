package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	playlistv1 "github.com/faqears/faqears/gen/go/playlist/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/go-chi/chi/v5"
)

type PlaylistHandler struct {
	c playlistv1.PlaylistServiceClient
}

func NewPlaylistHandler(c playlistv1.PlaylistServiceClient) *PlaylistHandler {
	return &PlaylistHandler{c: c}
}

func (h *PlaylistHandler) CreatePlaylist(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Public      bool   `json:"public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.CreatePlaylist(ctx, &playlistv1.CreatePlaylistRequest{
		OwnerId:     identity.UserID,
		Name:        body.Name,
		Description: body.Description,
		Public:      body.Public,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp.Playlist)
}

func (h *PlaylistHandler) ListUserPlaylists(w http.ResponseWriter, r *http.Request) {
	identity := callerIdentity(r.Context())
	q := r.URL.Query()
	limit := int32(20)
	offset := int32(0)
	inclCollab := false
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
	if v := q.Get("include_collaborations"); v == "true" || v == "1" {
		inclCollab = true
	}
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.ListUserPlaylists(ctx, &playlistv1.ListUserPlaylistsRequest{
		UserId:                identity.UserID,
		Limit:                 limit,
		Offset:                offset,
		IncludeCollaborations: inclCollab,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"playlists": resp.Playlists})
}

func (h *PlaylistHandler) GetPlaylist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.GetPlaylist(ctx, &playlistv1.GetPlaylistRequest{PlaylistId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Playlist)
}

func (h *PlaylistHandler) GetPlaylistByPermalink(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.GetPlaylistByPermalink(ctx, &playlistv1.GetPlaylistByPermalinkRequest{Permalink: code})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Playlist)
}

func (h *PlaylistHandler) UpdatePlaylist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Public      bool   `json:"public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	ctx := grpcx.OutgoingIdentity(r.Context(), callerIdentity(r.Context()))
	resp, err := h.c.UpdatePlaylist(ctx, &playlistv1.UpdatePlaylistRequest{
		PlaylistId:  id,
		Name:        body.Name,
		Description: body.Description,
		Public:      body.Public,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Playlist)
}

func (h *PlaylistHandler) DeletePlaylist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	_, err := h.c.DeletePlaylist(ctx, &playlistv1.DeletePlaylistRequest{
		PlaylistId: id,
		CallerId:   identity.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PlaylistHandler) AddTrack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		TrackID      string `json:"track_id"`
		AfterTrackID string `json:"after_track_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.AddTrack(ctx, &playlistv1.AddTrackRequest{
		PlaylistId:   id,
		TrackId:      body.TrackID,
		AfterTrackId: body.AfterTrackID,
		CallerId:     identity.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Playlist)
}

func (h *PlaylistHandler) RemoveTrack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	trackID := chi.URLParam(r, "track_id")
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.RemoveTrack(ctx, &playlistv1.RemoveTrackRequest{
		PlaylistId: id,
		TrackId:    trackID,
		CallerId:   identity.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Playlist)
}

func (h *PlaylistHandler) ReorderTrack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	trackID := chi.URLParam(r, "track_id")
	var body struct {
		AfterTrackID string `json:"after_track_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.c.ReorderTrack(ctx, &playlistv1.ReorderTrackRequest{
		PlaylistId:   id,
		TrackId:      trackID,
		AfterTrackId: body.AfterTrackID,
		CallerId:     identity.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Playlist)
}

func (h *PlaylistHandler) AddCollaborator(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var body struct {
		CollaboratorID string `json:"collaborator_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	_, err := h.c.AddCollaborator(ctx, &playlistv1.AddCollaboratorRequest{
		PlaylistId:     id,
		CollaboratorId: body.CollaboratorID,
		CallerId:       identity.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}

func (h *PlaylistHandler) RemoveCollaborator(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "user_id")
	identity := callerIdentity(r.Context())
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	_, err := h.c.RemoveCollaborator(ctx, &playlistv1.RemoveCollaboratorRequest{
		PlaylistId:     id,
		CollaboratorId: userID,
		CallerId:       identity.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{})
}
