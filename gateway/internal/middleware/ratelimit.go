package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func RateLimit(rdb *redis.Client, anonLimit, authedLimit int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			userID := GetUserID(ctx)

			var bucket, who string
			var limit int

			if userID != "" {
				bucket = "authed"
				who = userID
				limit = authedLimit
			} else {
				bucket = "anon"
				who = remoteIP(r)
				limit = anonLimit
			}

			minute := time.Now().UTC().Format("200601021504")
			key := fmt.Sprintf("gw:rl:%s:%s:%s", bucket, who, minute)

			count, err := rdb.Incr(ctx, key).Result()
			if err == nil && count == 1 {
				rdb.Expire(ctx, key, 2*time.Minute)
			}

			if err == nil && int(count) > limit {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"rate limit exceeded"}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func remoteIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		return ip[:idx]
	}
	return ip
}
