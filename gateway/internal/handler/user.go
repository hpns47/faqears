package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/faqears/faqears/gateway/internal/middleware"
	userv1 "github.com/faqears/faqears/gen/go/user/v1"
	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	user userv1.UserServiceClient
}

func NewUserHandler(user userv1.UserServiceClient) *UserHandler {
	return &UserHandler{user: user}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	resp, err := h.user.GetUser(r.Context(), &userv1.GetUserRequest{UserId: id})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.User)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	callerID := middleware.GetUserID(r.Context())
	if id != callerID {
		writeError(w, http.StatusForbidden, "cannot modify another user's profile")
		return
	}

	var body struct {
		DisplayName string `json:"display_name"`
		AvatarURL   string `json:"avatar_url"`
		Country     string `json:"country"`
		Language    string `json:"language"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.user.UpdateProfile(r.Context(), &userv1.UpdateProfileRequest{
		UserId:      id,
		DisplayName: body.DisplayName,
		AvatarUrl:   body.AvatarURL,
		Country:     body.Country,
		Language:    body.Language,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.User)
}

func (h *UserHandler) UpdateTier(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	callerID := middleware.GetUserID(r.Context())
	if id != callerID {
		writeError(w, http.StatusForbidden, "cannot change another user's tier")
		return
	}
	var body struct {
		Tier string `json:"tier"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if body.Tier != "free" && body.Tier != "premium" {
		writeError(w, http.StatusBadRequest, "tier must be 'free' or 'premium'")
		return
	}
	resp, err := h.user.UpdateTier(r.Context(), &userv1.UpdateTierRequest{
		UserId: id,
		Tier:   body.Tier,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp.User)
}

func (h *UserHandler) Follow(w http.ResponseWriter, r *http.Request) {
	followeeID := chi.URLParam(r, "id")
	callerID := middleware.GetUserID(r.Context())
	_, err := h.user.FollowUser(r.Context(), &userv1.FollowUserRequest{
		FollowerId: callerID,
		FolloweeId: followeeID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) Unfollow(w http.ResponseWriter, r *http.Request) {
	followeeID := chi.URLParam(r, "id")
	callerID := middleware.GetUserID(r.Context())
	_, err := h.user.UnfollowUser(r.Context(), &userv1.UnfollowUserRequest{
		FollowerId: callerID,
		FolloweeId: followeeID,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) ListFollowers(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit, offset := parsePagination(r)
	resp, err := h.user.ListFollowers(r.Context(), &userv1.ListFollowersRequest{
		UserId: id,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *UserHandler) ListFollowing(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	limit, offset := parsePagination(r)
	resp, err := h.user.ListFollowing(r.Context(), &userv1.ListFollowingRequest{
		UserId: id,
		Limit:  int32(limit),
		Offset: int32(offset),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func parsePagination(r *http.Request) (limit, offset int) {
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
