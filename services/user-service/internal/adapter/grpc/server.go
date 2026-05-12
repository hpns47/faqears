package grpc

import (
	"context"

	userv1 "github.com/faqears/faqears/gen/go/user/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/user-service/internal/domain"
	"github.com/faqears/faqears/services/user-service/internal/usecase"
)

type Server struct {
	userv1.UnimplementedUserServiceServer
	uc *usecase.User
}

func NewServer(uc *usecase.User) *Server {
	return &Server{uc: uc}
}

func (s *Server) GetUser(ctx context.Context, req *userv1.GetUserRequest) (*userv1.GetUserResponse, error) {
	u, err := s.uc.Get(ctx, req.GetUserId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &userv1.GetUserResponse{User: toProto(u)}, nil
}

func (s *Server) UpdateProfile(ctx context.Context, req *userv1.UpdateProfileRequest) (*userv1.UpdateProfileResponse, error) {
	u, err := s.uc.UpdateProfile(ctx, req.GetUserId(), domain.Profile{
		DisplayName: req.GetDisplayName(),
		AvatarURL:   req.GetAvatarUrl(),
		Country:     req.GetCountry(),
		Language:    req.GetLanguage(),
	})
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &userv1.UpdateProfileResponse{User: toProto(u)}, nil
}

func (s *Server) FollowUser(ctx context.Context, req *userv1.FollowUserRequest) (*userv1.FollowUserResponse, error) {
	if err := s.uc.Follow(ctx, req.GetFollowerId(), req.GetFolloweeId()); err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &userv1.FollowUserResponse{}, nil
}

func (s *Server) UnfollowUser(ctx context.Context, req *userv1.UnfollowUserRequest) (*userv1.UnfollowUserResponse, error) {
	if err := s.uc.Unfollow(ctx, req.GetFollowerId(), req.GetFolloweeId()); err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &userv1.UnfollowUserResponse{}, nil
}

func (s *Server) ListFollowers(ctx context.Context, req *userv1.ListFollowersRequest) (*userv1.ListFollowersResponse, error) {
	users, err := s.uc.ListFollowers(ctx, req.GetUserId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	out := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		out = append(out, toProto(u))
	}
	return &userv1.ListFollowersResponse{Followers: out}, nil
}

func (s *Server) ListFollowing(ctx context.Context, req *userv1.ListFollowingRequest) (*userv1.ListFollowingResponse, error) {
	users, err := s.uc.ListFollowing(ctx, req.GetUserId(), int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	out := make([]*userv1.User, 0, len(users))
	for _, u := range users {
		out = append(out, toProto(u))
	}
	return &userv1.ListFollowingResponse{Following: out}, nil
}

func (s *Server) UpdateTier(ctx context.Context, req *userv1.UpdateTierRequest) (*userv1.UpdateTierResponse, error) {
	u, err := s.uc.UpdateTier(ctx, req.GetUserId(), req.GetTier())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &userv1.UpdateTierResponse{User: toProto(u)}, nil
}

func toProto(u *domain.User) *userv1.User {
	return &userv1.User{
		Id:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		AvatarUrl:   u.AvatarURL,
		Country:     u.Country,
		Language:    u.Language,
		Tier:        u.Tier,
		CreatedAt:   u.CreatedAt.Unix(),
		UpdatedAt:   u.UpdatedAt.Unix(),
	}
}
