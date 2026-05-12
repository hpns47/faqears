package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	recov1 "github.com/faqears/faqears/gen/go/recommendation/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/pkg/postgres"
	pkgredis "github.com/faqears/faqears/pkg/redis"
	grpcadapter "github.com/faqears/faqears/services/recommendation-service/internal/adapter/grpc"
	kafkaadapter "github.com/faqears/faqears/services/recommendation-service/internal/adapter/kafka"
	pgadapter "github.com/faqears/faqears/services/recommendation-service/internal/adapter/postgres"
	redisadapter "github.com/faqears/faqears/services/recommendation-service/internal/adapter/redis"
	"github.com/faqears/faqears/services/recommendation-service/internal/config"
	"github.com/faqears/faqears/services/recommendation-service/internal/job"
	"github.com/faqears/faqears/services/recommendation-service/internal/usecase"
)

func main() {
	log := logger.New("recommendation-service", os.Getenv("RECOMMENDATION_LOG_LEVEL"), os.Stdout)
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
	log = logger.New("recommendation-service", cfg.LogLevel, os.Stdout)

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

	playRepo := pgadapter.NewPlayRepo(pool)
	simRepo := pgadapter.NewSimilarityRepo(pool)
	cache := redisadapter.NewCache(redisClient)

	uc := usecase.NewRecommendation(playRepo, simRepo, cache, log)

	consumer := kafkaadapter.NewPlayEventsConsumer(cfg.Brokers, cfg.PlayTopic, cfg.GroupID, playRepo, log)
	defer consumer.Close()

	refresher := job.NewRefresher(playRepo, simRepo, cache, cfg.RefreshInterval, log)

	srv, lis, err := grpcx.NewServer(grpcx.ServerConfig{Addr: cfg.GRPCAddr, Logger: log})
	if err != nil {
		return err
	}
	recov1.RegisterRecommendationServiceServer(srv, grpcadapter.NewServer(uc))

	errCh := make(chan error, 2)
	go func() {
		log.Info("grpc server listening", slog.String("addr", cfg.GRPCAddr))
		errCh <- srv.Serve(lis)
	}()
	go func() {
		log.Info("kafka consumer started", slog.String("topic", cfg.PlayTopic))
		errCh <- consumer.Run(ctx)
	}()
	go func() {
		log.Info("similarity refresher started", slog.String("interval", cfg.RefreshInterval.String()))
		refresher.Run(ctx)
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
