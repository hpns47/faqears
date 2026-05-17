package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/faqears/faqears/gateway/internal/middleware"
	authv1 "github.com/faqears/faqears/gen/go/auth/v1"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	auth authv1.AuthServiceClient
}

func NewAuthHandler(auth authv1.AuthServiceClient) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req authv1.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.auth.Register(r.Context(), &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req authv1.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.UserAgent = r.UserAgent()
	resp, err := h.auth.Login(r.Context(), &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req authv1.RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	req.UserAgent = r.UserAgent()
	resp, err := h.auth.RefreshToken(r.Context(), &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	_, err := h.auth.ChangePassword(r.Context(), &authv1.ChangePasswordRequest{
		UserId:      userID,
		OldPassword: body.OldPassword,
		NewPassword: body.NewPassword,
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	resp, err := h.auth.ListSessions(r.Context(), &authv1.ListSessionsRequest{
		UserId:              userID,
		CurrentRefreshToken: r.Header.Get("X-Current-Refresh-Token"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": resp.Sessions})
}

func (h *AuthHandler) RevokeSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	_, err := h.auth.RevokeSession(r.Context(), &authv1.RevokeSessionRequest{
		UserId:    userID,
		SessionId: chi.URLParam(r, "id"),
	})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	_, err := h.auth.Logout(r.Context(), &authv1.LogoutRequest{AccessToken: token})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) ValidateToken(w http.ResponseWriter, r *http.Request) {
	header := r.Header.Get("Authorization")
	token := strings.TrimPrefix(header, "Bearer ")
	resp, err := h.auth.ValidateToken(r.Context(), &authv1.ValidateTokenRequest{AccessToken: token})
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user_id": resp.UserId,
		"email":   resp.Email,
		"roles":   resp.Roles,
	})
}

func (h *AuthHandler) OAuthAuthorizeURL(w http.ResponseWriter, r *http.Request) {
	var req authv1.OAuthAuthorizeURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.auth.OAuthAuthorizeURL(r.Context(), &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"authorize_url": resp.AuthorizeUrl,
		"state":         resp.State,
	})
}

func (h *AuthHandler) OAuthCallback(w http.ResponseWriter, r *http.Request) {
	var req authv1.OAuthCallbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	resp, err := h.auth.OAuthCallback(r.Context(), &req)
	if err != nil {
		writeGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"access_token":     resp.AccessToken,
		"refresh_token":    resp.RefreshToken,
		"access_expires_at": resp.AccessExpiresAt,
		"user_id":          resp.UserId,
		"created":          resp.Created,
	})
}
