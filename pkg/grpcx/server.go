package grpcx

import (
	"context"
	"log/slog"
	"net"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

type ServerConfig struct {
	Addr       string
	Logger     *slog.Logger
	Extra      []grpc.UnaryServerInterceptor
}

func NewServer(cfg ServerConfig) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", cfg.Addr)
	if err != nil {
		return nil, nil, err
	}
	chain := []grpc.UnaryServerInterceptor{
		RecoveryInterceptor(cfg.Logger),
		LoggingInterceptor(cfg.Logger),
		IdentityInterceptor(),
	}
	chain = append(chain, cfg.Extra...)
	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(chain...))
	hs := health.NewServer()
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(srv, hs)
	reflection.Register(srv)
	return srv, lis, nil
}

func RecoveryInterceptor(l *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if r := recover(); r != nil {
				l.ErrorContext(ctx, "panic recovered",
					slog.String("method", info.FullMethod),
					slog.Any("panic", r),
					slog.String("stack", string(debug.Stack())),
				)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

func LoggingInterceptor(l *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		code := status.Code(err)
		attrs := []any{
			slog.String("method", info.FullMethod),
			slog.String("code", code.String()),
			slog.Duration("duration", time.Since(start)),
		}
		if err != nil {
			attrs = append(attrs, slog.String("error", err.Error()))
			l.WarnContext(ctx, "rpc finished", attrs...)
		} else {
			l.InfoContext(ctx, "rpc finished", attrs...)
		}
		return resp, err
	}
}
