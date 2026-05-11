package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/faqears/faqears/gateway/internal/grpcclients"
	authv1 "github.com/faqears/faqears/gen/go/auth/v1"
	catalogv1 "github.com/faqears/faqears/gen/go/catalog/v1"
	userv1 "github.com/faqears/faqears/gen/go/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type HealthHandler struct {
	clients *grpcclients.Clients
}

func NewHealthHandler(clients *grpcclients.Clients) *HealthHandler {
	return &HealthHandler{clients: clients}
}

func (h *HealthHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
	defer cancel()

	authOk := pingAuth(ctx, h.clients.Auth)
	userOk := pingUser(ctx, h.clients.User)
	catalogOk := pingCatalog(ctx, h.clients.Catalog)

	result := map[string]string{
		"auth":    boolStatus(authOk),
		"user":    boolStatus(userOk),
		"catalog": boolStatus(catalogOk),
	}

	if authOk && userOk && catalogOk {
		writeJSON(w, http.StatusOK, result)
	} else {
		writeJSON(w, http.StatusServiceUnavailable, result)
	}
}

func boolStatus(ok bool) string {
	if ok {
		return "ok"
	}
	return "unreachable"
}

func upstreamReachable(err error) bool {
	if err == nil {
		return true
	}
	st, _ := grpcstatus.FromError(err)
	return st.Code() != codes.Unavailable
}

func pingAuth(ctx context.Context, c authv1.AuthServiceClient) bool {
	_, err := c.ValidateToken(ctx, &authv1.ValidateTokenRequest{AccessToken: ""}, grpc.WaitForReady(false))
	return upstreamReachable(err)
}

func pingUser(ctx context.Context, c userv1.UserServiceClient) bool {
	_, err := c.GetUser(ctx, &userv1.GetUserRequest{UserId: ""}, grpc.WaitForReady(false))
	return upstreamReachable(err)
}

func pingCatalog(ctx context.Context, c catalogv1.CatalogServiceClient) bool {
	_, err := c.GetTrack(ctx, &catalogv1.GetTrackRequest{TrackId: ""}, grpc.WaitForReady(false))
	return upstreamReachable(err)
}
