package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	catalogv1 "github.com/faqears/faqears/gen/go/catalog/v1"
	"google.golang.org/grpc"

	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/pkg/postgres"
	pkgredis "github.com/faqears/faqears/pkg/redis"
	grpcadapter "github.com/faqears/faqears/services/catalog-service/internal/adapter/grpc"
	kafkaadapter "github.com/faqears/faqears/services/catalog-service/internal/adapter/kafka"
	pgadapter "github.com/faqears/faqears/services/catalog-service/internal/adapter/postgres"
	redisadapter "github.com/faqears/faqears/services/catalog-service/internal/adapter/redis"
	"github.com/faqears/faqears/services/catalog-service/internal/config"
	"github.com/faqears/faqears/services/catalog-service/internal/usecase"
)

func main() {
	log := logger.New("catalog-service", os.Getenv("CATALOG_LOG_LEVEL"), os.Stdout)
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
	log = logger.New("catalog-service", cfg.LogLevel, os.Stdout)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := postgres.Migrate(cfg.DBURL, cfg.MigrateURL); err != nil {
		return err
	}
	log.Info("migrations applied")

	pool, err := postgres.NewPool(ctx, postgres.Config{
		URL:             cfg.DBURL,
		MaxConns:        20,
		MinConns:        2,
		MaxConnLifetime: time.Hour,
		MaxConnIdleTime: 30 * time.Minute,
	})
	if err != nil {
		return err
	}
	defer pool.Close()
	log.Info("postgres connected")

	redisClient, err := pkgredis.New(ctx, pkgredis.Config{Addr: cfg.RedisAddr})
	if err != nil {
		return err
	}
	defer redisClient.Close()
	log.Info("redis connected")

	publisher := kafkaadapter.NewPublisher(cfg.Brokers, cfg.EventTopic)
	defer publisher.Close()

	catalog := usecase.NewCatalog(
		pgadapter.NewArtistRepo(pool),
		pgadapter.NewAlbumRepo(pool),
		pgadapter.NewTrackRepo(pool),
		redisadapter.NewCache(redisClient),
		publisher,
		cfg.CacheTTL,
	)

	adminGuard := grpcx.RequireMethodAnyRole(map[string][]string{
		"/catalog.v1.CatalogService/IngestArtist": {"admin", "service"},
		"/catalog.v1.CatalogService/IngestAlbum":  {"admin", "service"},
		"/catalog.v1.CatalogService/IngestTrack":  {"admin", "service", "user"},
	})
	srv, lis, err := grpcx.NewServer(grpcx.ServerConfig{
		Addr:   cfg.GRPCAddr,
		Logger: log,
		Extra:  []grpc.UnaryServerInterceptor{adminGuard},
	})
	if err != nil {
		return err
	}
	catalogv1.RegisterCatalogServiceServer(srv, grpcadapter.NewServer(catalog))

	errCh := make(chan error, 1)
	go func() {
		log.Info("grpc server listening", slog.String("addr", cfg.GRPCAddr))
		errCh <- srv.Serve(lis)
	}()

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
	}

	done := make(chan struct{})
	go func() {
		srv.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
		log.Info("grpc server stopped")
	case <-time.After(15 * time.Second):
		srv.Stop()
		log.Warn("grpc server force stopped")
	}
	return nil
}
