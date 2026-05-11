package server

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/faqears/faqears/gateway/internal/grpcclients"
	"github.com/faqears/faqears/gateway/internal/handler"
	"github.com/faqears/faqears/gateway/internal/middleware"
	authv1 "github.com/faqears/faqears/gen/go/auth/v1"
	"github.com/go-chi/chi/v5"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	Addr            string
	TokenCacheTTL   time.Duration
	RateLimitAnon   int
	RateLimitAuthed int
}

func New(cfg Config, clients *grpcclients.Clients, rdb *redis.Client, log *slog.Logger) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(log))
	r.Use(middleware.Recovery(log))
	r.Use(middleware.CORS)
	r.Use(middleware.RateLimit(rdb, cfg.RateLimitAnon, cfg.RateLimitAuthed))

	authH := handler.NewAuthHandler(clients.Auth)
	userH := handler.NewUserHandler(clients.User)
	catalogH := handler.NewCatalogHandler(clients.Catalog)
	healthH := handler.NewHealthHandler(clients)

	r.Get("/healthz", healthH.Healthz)
	r.Get("/readyz", healthH.Readyz)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.Refresh)

			r.Group(func(r chi.Router) {
				r.Use(authMW(clients.Auth, rdb, cfg.TokenCacheTTL))
				r.Post("/logout", authH.Logout)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(authMW(clients.Auth, rdb, cfg.TokenCacheTTL))

			r.Route("/users/{id}", func(r chi.Router) {
				r.Get("/", userH.GetUser)
				r.Patch("/", userH.UpdateUser)
				r.Post("/follow", userH.Follow)
				r.Delete("/follow", userH.Unfollow)
				r.Get("/followers", userH.ListFollowers)
				r.Get("/following", userH.ListFollowing)
			})

			r.Get("/tracks/{id}", catalogH.GetTrack)
			r.Get("/albums/{id}", catalogH.GetAlbum)
			r.Get("/artists/{id}", catalogH.GetArtist)
			r.Get("/albums/{id}/tracks", catalogH.ListTracksByAlbum)
			r.Get("/artists/{id}/albums", catalogH.ListAlbumsByArtist)
			r.Get("/search", catalogH.Search)

			r.Route("/admin/catalog", func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Post("/artists", catalogH.IngestArtist)
				r.Post("/albums", catalogH.IngestAlbum)
				r.Post("/tracks", catalogH.IngestTrack)
			})
		})
	})

	return &http.Server{
		Addr:         cfg.Addr,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
}

func authMW(authClient authv1.AuthServiceClient, rdb *redis.Client, ttl time.Duration) func(http.Handler) http.Handler {
	return middleware.Auth(authClient, rdb, ttl)
}
