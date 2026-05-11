package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	authv1 "github.com/faqears/faqears/gen/go/auth/v1"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/pkg/postgres"
	pkgredis "github.com/faqears/faqears/pkg/redis"
	"github.com/faqears/faqears/services/auth-service/internal/adapter/bcrypt"
	grpcadapter "github.com/faqears/faqears/services/auth-service/internal/adapter/grpc"
	"github.com/faqears/faqears/services/auth-service/internal/adapter/jwt"
	kafkaadapter "github.com/faqears/faqears/services/auth-service/internal/adapter/kafka"
	oauthadapter "github.com/faqears/faqears/services/auth-service/internal/adapter/oauth"
	redisadapter "github.com/faqears/faqears/services/auth-service/internal/adapter/redis"
	pgadapter "github.com/faqears/faqears/services/auth-service/internal/adapter/postgres"
	"github.com/faqears/faqears/services/auth-service/internal/config"
	"github.com/faqears/faqears/services/auth-service/internal/port"
	"github.com/faqears/faqears/services/auth-service/internal/usecase"
)

func main() {
	log := logger.New("auth-service", os.Getenv("AUTH_LOG_LEVEL"), os.Stdout)
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
	log = logger.New("auth-service", cfg.LogLevel, os.Stdout)

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

	userRepo := pgadapter.NewUserRepo(pool)
	tokenRepo := pgadapter.NewTokenRepo(pool)
	hasher := bcrypt.New(cfg.BcryptCost)
	issuer := jwt.New(jwt.Config{
		Secret:     cfg.JWTSecret,
		Issuer:     cfg.JWTIssuer,
		AccessTTL:  cfg.AccessTTL,
		RefreshTTL: cfg.RefreshTTL,
	})

	auth := usecase.NewAuth(userRepo, tokenRepo, hasher, issuer, publisher)

	google := oauthadapter.NewGoogle(oauthadapter.GoogleConfig{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
	})
	github := oauthadapter.NewGitHub(oauthadapter.GitHubConfig{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
		RedirectURL:  cfg.GitHubRedirectURL,
	})
	registry := oauthadapter.NewRegistry(google, github)

	var stateStore port.OAuthStateStore
	if cfg.RedisAddr != "" {
		rc, rerr := pkgredis.New(ctx, pkgredis.Config{Addr: cfg.RedisAddr})
		if rerr != nil {
			return rerr
		}
		defer rc.Close()
		stateStore = redisadapter.NewStateStore(rc)
		log.Info("redis connected for oauth state store", slog.String("addr", cfg.RedisAddr))
	} else {
		log.Warn("AUTH_REDIS_ADDR not set — oauth state store disabled, OAuthAuthorizeURL and OAuthCallback will return unavailable")
		stateStore = redisadapter.NewNoopStateStore()
	}

	oauthUC := usecase.NewOAuth(registry, stateStore, cfg.OAuthStateTTL, userRepo, tokenRepo, hasher, issuer, publisher)

	srv, lis, err := grpcx.NewServer(grpcx.ServerConfig{Addr: cfg.GRPCAddr, Logger: log})
	if err != nil {
		return err
	}
	authv1.RegisterAuthServiceServer(srv, grpcadapter.NewServer(auth, oauthUC))

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
