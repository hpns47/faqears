package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	userv1 "github.com/faqears/faqears/gen/go/user/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/pkg/postgres"
	grpcadapter "github.com/faqears/faqears/services/user-service/internal/adapter/grpc"
	kafkaadapter "github.com/faqears/faqears/services/user-service/internal/adapter/kafka"
	pgadapter "github.com/faqears/faqears/services/user-service/internal/adapter/postgres"
	"github.com/faqears/faqears/services/user-service/internal/config"
	"github.com/faqears/faqears/services/user-service/internal/usecase"
)

func main() {
	log := logger.New("user-service", os.Getenv("USER_LOG_LEVEL"), os.Stdout)
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
	log = logger.New("user-service", cfg.LogLevel, os.Stdout)

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

	publisher := kafkaadapter.NewPublisher(cfg.Brokers, cfg.EventTopic)
	defer publisher.Close()

	uc := usecase.NewUser(
		pgadapter.NewUserRepo(pool),
		pgadapter.NewFollowRepo(pool),
		publisher,
	)

	consumer := kafkaadapter.NewAuthEventsConsumer(cfg.Brokers, cfg.AuthTopic, cfg.GroupID, uc, log)
	defer consumer.Close()

	srv, lis, err := grpcx.NewServer(grpcx.ServerConfig{Addr: cfg.GRPCAddr, Logger: log})
	if err != nil {
		return err
	}
	userv1.RegisterUserServiceServer(srv, grpcadapter.NewServer(uc))

	errCh := make(chan error, 2)
	go func() {
		log.Info("grpc server listening", slog.String("addr", cfg.GRPCAddr))
		errCh <- srv.Serve(lis)
	}()
	go func() {
		log.Info("kafka consumer started", slog.String("topic", cfg.AuthTopic))
		errCh <- consumer.Run(ctx)
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
