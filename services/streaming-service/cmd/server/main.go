package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	streamingv1 "github.com/faqears/faqears/gen/go/streaming/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/pkg/logger"
	pkgminio "github.com/faqears/faqears/pkg/minio"
	"github.com/faqears/faqears/pkg/postgres"
	pkgredis "github.com/faqears/faqears/pkg/redis"
	grpcadapter "github.com/faqears/faqears/services/streaming-service/internal/adapter/grpc"
	kafkaadapter "github.com/faqears/faqears/services/streaming-service/internal/adapter/kafka"
	minioadapter "github.com/faqears/faqears/services/streaming-service/internal/adapter/minio"
	pgadapter "github.com/faqears/faqears/services/streaming-service/internal/adapter/postgres"
	redisadapter "github.com/faqears/faqears/services/streaming-service/internal/adapter/redis"
	"github.com/faqears/faqears/services/streaming-service/internal/config"
	"github.com/faqears/faqears/services/streaming-service/internal/usecase"
)

func main() {
	log := logger.New("streaming-service", os.Getenv("STREAMING_LOG_LEVEL"), os.Stdout)
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
	log = logger.New("streaming-service", cfg.LogLevel, os.Stdout)

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

	minioClient, err := pkgminio.New(ctx, pkgminio.Config{
		Endpoint:  cfg.MinioEndpoint,
		AccessKey: cfg.MinioAccessKey,
		SecretKey: cfg.MinioSecretKey,
		UseSSL:    cfg.MinioUseSSL,
		Region:    cfg.MinioRegion,
	})
	if err != nil {
		return err
	}
	if err := minioClient.EnsureBuckets(ctx, cfg.MinioBucket); err != nil {
		return err
	}
	log.Info("minio connected", slog.String("bucket", cfg.MinioBucket))

	presignClient, err := pkgminio.NewUnverified(pkgminio.Config{
		Endpoint:  cfg.MinioPublicEndpoint,
		AccessKey: cfg.MinioAccessKey,
		SecretKey: cfg.MinioSecretKey,
		UseSSL:    cfg.MinioPublicUseSSL,
		Region:    cfg.MinioRegion,
	})
	if err != nil {
		return err
	}
	log.Info("minio presign endpoint", slog.String("endpoint", cfg.MinioPublicEndpoint))

	publisher := kafkaadapter.NewPublisher(cfg.Brokers, cfg.KafkaTopic)
	defer publisher.Close()

	uc := usecase.NewStreaming(
		pgadapter.NewAudioRepo(pool),
		redisadapter.NewSessionStore(redisClient),
		minioadapter.NewStorage(minioClient, presignClient, cfg.MinioBucket),
		publisher,
		cfg.MinioBucket,
		cfg.SignedURLTTL,
	)

	srv, lis, err := grpcx.NewServer(grpcx.ServerConfig{Addr: cfg.GRPCAddr, Logger: log})
	if err != nil {
		return err
	}
	streamingv1.RegisterStreamingServiceServer(srv, grpcadapter.NewServer(uc))

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
