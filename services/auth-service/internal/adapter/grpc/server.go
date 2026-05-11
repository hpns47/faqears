package grpc

import (
	"context"

	authv1 "github.com/faqears/faqears/gen/go/auth/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/auth-service/internal/usecase"
)

type Server struct {
	authv1.UnimplementedAuthServiceServer
	auth  *usecase.Auth
	oauth *usecase.OAuth
}

func NewServer(auth *usecase.Auth, oauth *usecase.OAuth) *Server {
	return &Server{auth: auth, oauth: oauth}
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

func (s *Server) OAuthAuthorizeURL(ctx context.Context, req *authv1.OAuthAuthorizeURLRequest) (*authv1.OAuthAuthorizeURLResponse, error) {
	authorizeURL, state, err := s.oauth.AuthorizeURL(ctx, req.GetProvider(), req.GetRedirectUri())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &authv1.OAuthAuthorizeURLResponse{
		AuthorizeUrl: authorizeURL,
		State:        state,
	}, nil
}

func (s *Server) OAuthCallback(ctx context.Context, req *authv1.OAuthCallbackRequest) (*authv1.OAuthCallbackResponse, error) {
	result, err := s.oauth.Callback(ctx, req.GetProvider(), req.GetCode(), req.GetState())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &authv1.OAuthCallbackResponse{
		AccessToken:     result.AccessToken,
		RefreshToken:    result.RefreshToken,
		AccessExpiresAt: result.AccessExpiresAt.Unix(),
		UserId:          result.UserID,
		Created:         result.Created,
	}, nil
}
