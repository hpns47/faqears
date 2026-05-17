package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/faqears/faqears/gateway/internal/config"
	"github.com/faqears/faqears/gateway/internal/grpcclients"
	"github.com/faqears/faqears/gateway/internal/server"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/redis/go-redis/v9"
)

func main() {
	log := logger.New("api-gateway", os.Getenv("GATEWAY_LOG_LEVEL"), os.Stdout)
	slog.SetDefault(log)

	if err := run(log); err != nil {
		log.Error("service stopped with error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log = logger.New("api-gateway", cfg.LogLevel, os.Stdout)

	clients, err := grpcclients.New(cfg.AuthGRPCAddr, cfg.UserGRPCAddr, cfg.CatalogGRPCAddr, cfg.StreamingGRPCAddr, cfg.PlaylistGRPCAddr, cfg.RecommendationGRPCAddr, cfg.PaymentGRPCAddr, cfg.GenerationGRPCAddr)
	if err != nil {
		return err
	}
	defer clients.Close()
	log.Info("grpc clients connected")

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()
	log.Info("redis connected", slog.String("addr", cfg.RedisAddr))

	srv := server.New(server.Config{
		Addr:            cfg.HTTPAddr,
		TokenCacheTTL:   cfg.TokenCacheTTL,
		RateLimitAnon:   cfg.RateLimitAnon,
		RateLimitAuthed: cfg.RateLimitAuthed,
	}, clients, rdb, log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info("http server listening", slog.String("addr", cfg.HTTPAddr))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutCtx); err != nil {
		log.Warn("http server shutdown error", slog.String("error", err.Error()))
	}
	log.Info("http server stopped")
	return nil
}
