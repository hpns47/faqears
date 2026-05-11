package grpc

import (
	"context"

	authv1 "github.com/faqears/faqears/gen/go/auth/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/auth-service/internal/usecase"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer
	auth *usecase.Auth
}

func NewServer(auth *usecase.Auth) *Server {
	return &Server{auth: auth}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	id, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &authv1.RegisterResponse{UserId: id}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	pair, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &authv1.LoginResponse{
		AccessToken:     pair.AccessToken,
		RefreshToken:    pair.RefreshToken,
		AccessExpiresAt: pair.AccessExpiresAt.Unix(),
	}, nil
}

func (s *Server) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	pair, err := s.auth.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &authv1.RefreshTokenResponse{
		AccessToken:     pair.AccessToken,
		RefreshToken:    pair.RefreshToken,
		AccessExpiresAt: pair.AccessExpiresAt.Unix(),
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := s.auth.ValidateToken(ctx, req.GetAccessToken())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &authv1.ValidateTokenResponse{
		UserId: claims.UserID,
		Email:  claims.Email,
		Roles:  claims.Roles,
	}, nil
}

func (s *Server) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	if err := s.auth.Logout(ctx, req.GetAccessToken()); err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &authv1.LogoutResponse{}, nil
}
