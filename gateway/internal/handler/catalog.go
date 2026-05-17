package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	catalogv1 "github.com/faqears/faqears/gen/go/catalog/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/go-chi/chi/v5"
)

type CatalogHandler struct {
	catalog catalogv1.CatalogServiceClient
}

func NewCatalogHandler(catalog catalogv1.CatalogServiceClient) *CatalogHandler {
	return &CatalogHandler{catalog: catalog}
}

func (h *CatalogHandler) GetTrack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.catalog.GetTrack(r.Context(), &catalogv1.GetTrackRequest{TrackId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Track)
}

func (h *CatalogHandler) GetAlbum(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.catalog.GetAlbum(r.Context(), &catalogv1.GetAlbumRequest{AlbumId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Album)
}

func (h *CatalogHandler) GetArtist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.catalog.GetArtist(r.Context(), &catalogv1.GetArtistRequest{ArtistId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.Artist)
}

func (h *CatalogHandler) ListTracksByAlbum(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit, offset := parseCatalogPagination(r)
	resp, err := h.catalog.ListTracksByAlbum(r.Context(), &catalogv1.ListTracksByAlbumRequest{
		AlbumId: id,
		Limit:   int32(limit),
		Offset:  int32(offset),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *CatalogHandler) ListAlbumsByArtist(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit, offset := parseCatalogPagination(r)
	resp, err := h.catalog.ListAlbumsByArtist(r.Context(), &catalogv1.ListAlbumsByArtistRequest{
		ArtistId: id,
		Limit:    int32(limit),
		Offset:   int32(offset),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *CatalogHandler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	limitStr := r.URL.Query().Get("limit")
	limit := int32(20)
	if limitStr != "" {
		if n, err := strconv.Atoi(limitStr); err == nil && n > 0 {
			limit = int32(n)
		}
	}
	resp, err := h.catalog.Search(r.Context(), &catalogv1.SearchRequest{
		Query: q,
		Limit: limit,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *CatalogHandler) IngestArtist(w http.ResponseWriter, r *http.Request) {
	var req catalogv1.IngestArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.catalog.IngestArtist(r.Context(), &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *CatalogHandler) IngestAlbum(w http.ResponseWriter, r *http.Request) {
	var req catalogv1.IngestAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.catalog.IngestAlbum(r.Context(), &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *CatalogHandler) IngestTrack(w http.ResponseWriter, r *http.Request) {
	var req catalogv1.IngestTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.catalog.IngestTrack(r.Context(), &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

const (
	systemArtistID = "11111111-1111-1111-1111-111111111111"
	systemAlbumID  = "22222222-2222-2222-2222-222222222222"
)

func (h *CatalogHandler) CreateUserTrack(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title       string   `json:"title"`
		DurationSec int32    `json:"duration_sec"`
		Genres      []string `json:"genres"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Title == "" {
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}
	identity := callerIdentity(r.Context())
	if identity.UserID == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	ctx := grpcx.OutgoingIdentity(r.Context(), identity)
	resp, err := h.catalog.IngestTrack(ctx, &catalogv1.IngestTrackRequest{
		AlbumId:       systemAlbumID,
		ArtistId:      systemArtistID,
		Title:         body.Title,
		DurationSec:   body.DurationSec,
		Genres:        body.Genres,
		UserGenerated: true,
		OwnerUserId:   identity.UserID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func parseCatalogPagination(r *http.Request) (limit, offset int) {
	limit = 20
	offset = 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	return
}
