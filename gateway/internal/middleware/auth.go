package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	authv1 "github.com/faqears/faqears/gen/go/auth/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/redis/go-redis/v9"
)

type cachedClaims struct {
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
}

func Auth(authClient authv1.AuthServiceClient, rdb *redis.Client, ttl time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"missing or invalid authorization header"}`))
				return
			}
			token := strings.TrimPrefix(header, "Bearer ")

			claims, err := resolveClaims(r.Context(), token, authClient, rdb, ttl)
			if err != nil || claims == nil || claims.UserID == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"error":"invalid or expired token"}`))
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
			ctx = grpcx.OutgoingIdentity(ctx, grpcx.Identity{
				UserID: claims.UserID,
				Email:  claims.Email,
				Roles:  claims.Roles,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func resolveClaims(ctx context.Context, token string, authClient authv1.AuthServiceClient, rdb *redis.Client, ttl time.Duration) (*cachedClaims, error) {
	h := sha256.Sum256([]byte(token))
	key := fmt.Sprintf("gw:token:%x", h)

	val, err := rdb.Get(ctx, key).Result()
	if err == nil {
		var claims cachedClaims
		if jsonErr := json.Unmarshal([]byte(val), &claims); jsonErr == nil && claims.UserID != "" {
			return &claims, nil
		}
	}

	resp, err := authClient.ValidateToken(ctx, &authv1.ValidateTokenRequest{AccessToken: token})
	if err != nil {
		return nil, err
	}
	if resp.UserId == "" {
		return nil, fmt.Errorf("empty user_id in token claims")
	}

	claims := &cachedClaims{
		UserID: resp.UserId,
		Email:  resp.Email,
		Roles:  resp.Roles,
	}

	if data, jsonErr := json.Marshal(claims); jsonErr == nil {
		rdb.Set(ctx, key, data, ttl)
	}

	return claims, nil
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles := GetUserRoles(r.Context())
			for _, rr := range roles {
				if rr == role {
					next.ServeHTTP(w, r)
					return
				}
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"error":"forbidden"}`))
		})
	}
}
