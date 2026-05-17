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
	streamingH := handler.NewStreamingHandler(clients.Streaming, clients.Catalog)
	playlistH := handler.NewPlaylistHandler(clients.Playlist)
	recommendationH := handler.NewRecommendationHandler(clients.Recommendation)
	paymentH := handler.NewPaymentHandler(clients.Payment)
	generationH := handler.NewGenerationHandler(clients.Generation)

	r.Get("/healthz", healthH.Healthz)
	r.Get("/readyz", healthH.Readyz)
	mountDocs(r)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authH.Register)
			r.Post("/login", authH.Login)
			r.Post("/refresh", authH.Refresh)
			r.Post("/oauth/authorize-url", authH.OAuthAuthorizeURL)
			r.Post("/oauth/callback", authH.OAuthCallback)

			r.Group(func(r chi.Router) {
				r.Use(authMW(clients.Auth, rdb, cfg.TokenCacheTTL))
				r.Post("/logout", authH.Logout)
				r.Post("/validate", authH.ValidateToken)
				r.Post("/change-password", authH.ChangePassword)
				r.Get("/sessions", authH.ListSessions)
				r.Delete("/sessions/{id}", authH.RevokeSession)
			})
		})

		r.Post("/payments/nowpayments/webhook", paymentH.HandleWebhook)

		r.Group(func(r chi.Router) {
			r.Use(authMW(clients.Auth, rdb, cfg.TokenCacheTTL))

			r.Route("/users/{id}", func(r chi.Router) {
				r.Get("/", userH.GetUser)
				r.Patch("/", userH.UpdateUser)
				r.Put("/tier", userH.UpdateTier)
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

			r.Route("/streaming", func(r chi.Router) {
				r.Get("/tracks/{id}/url", streamingH.GetStreamURL)
				r.Get("/tracks/{id}/stream", streamingH.RedirectToStream)
				r.Post("/play", streamingH.RegisterPlay)
				r.Post("/skip", streamingH.RegisterSkip)
				r.Post("/complete", streamingH.RegisterComplete)
				r.Post("/tracks/{id}/audio", streamingH.UploadUserAudio)
			})

			r.Post("/tracks", catalogH.CreateUserTrack)

			r.Route("/admin/streaming", func(r chi.Router) {
				r.Use(middleware.RequireRole("admin"))
				r.Post("/tracks/{id}/audio", streamingH.UploadAudio)
			})

			r.Route("/playlists", func(r chi.Router) {
				r.Post("/", playlistH.CreatePlaylist)
				r.Get("/", playlistH.ListUserPlaylists)
				r.Get("/by-permalink/{code}", playlistH.GetPlaylistByPermalink)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", playlistH.GetPlaylist)
					r.Patch("/", playlistH.UpdatePlaylist)
					r.Delete("/", playlistH.DeletePlaylist)
					r.Post("/tracks", playlistH.AddTrack)
					r.Delete("/tracks/{track_id}", playlistH.RemoveTrack)
					r.Patch("/tracks/{track_id}/rank", playlistH.ReorderTrack)
					r.Post("/collaborators", playlistH.AddCollaborator)
					r.Delete("/collaborators/{user_id}", playlistH.RemoveCollaborator)
				})
			})

			r.Route("/recommendations", func(r chi.Router) {
				r.Get("/daily-mix", recommendationH.DailyMix)
				r.Get("/discover-weekly", recommendationH.DiscoverWeekly)
				r.Get("/similar/{track_id}", recommendationH.SimilarTracks)
				r.Get("/top", recommendationH.TopTracks)
			})

			r.Route("/payments", func(r chi.Router) {
				r.Post("/invoices", paymentH.CreateInvoice)
				r.Get("/", paymentH.ListPayments)
				r.Get("/{id}", paymentH.GetPayment)
			})

			r.Route("/generation", func(r chi.Router) {
				r.Post("/lyrics", generationH.GenerateLyrics)
				r.Post("/lyrics/{job_id}/revise", generationH.ReviseLyrics)
				r.Post("/music", generationH.GenerateMusic)
				r.Route("/jobs", func(r chi.Router) {
					r.Get("/", generationH.ListJobs)
					r.Get("/{job_id}", generationH.GetJob)
					r.Post("/{job_id}/cancel", generationH.CancelJob)
					r.Get("/{job_id}/events", generationH.JobEvents)
				})
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
