package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	paymentv1 "github.com/faqears/faqears/gen/go/payment/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/pkg/postgres"
	grpcadapter "github.com/faqears/faqears/services/payment-service/internal/adapter/grpc"
	kafkaadapter "github.com/faqears/faqears/services/payment-service/internal/adapter/kafka"
	"github.com/faqears/faqears/services/payment-service/internal/adapter/nowpayments"
	pgadapter "github.com/faqears/faqears/services/payment-service/internal/adapter/postgres"
	"github.com/faqears/faqears/services/payment-service/internal/config"
	"github.com/faqears/faqears/services/payment-service/internal/usecase"
)

func main() {
	log := logger.New("payment-service", os.Getenv("PAYMENT_LOG_LEVEL"), os.Stdout)
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
	log = logger.New("payment-service", cfg.LogLevel, os.Stdout)

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

	publisher := kafkaadapter.NewPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	defer publisher.Close()

	paymentRepo := pgadapter.NewPaymentRepo(pool)
	eventRepo := pgadapter.NewEventRepo(pool)
	client := nowpayments.NewClient(cfg.NowPaymentsBaseURL, cfg.NowPaymentsAPIKey, cfg.NowPaymentsIPNSecret, cfg.IPNCallbackURL)

	uc := usecase.NewPayment(paymentRepo, eventRepo, client, publisher, log)

	srv, lis, err := grpcx.NewServer(grpcx.ServerConfig{Addr: cfg.GRPCAddr, Logger: log})
	if err != nil {
		return err
	}
	paymentv1.RegisterPaymentServiceServer(srv, grpcadapter.NewServer(uc))

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
