package port

import (
	"context"

	"github.com/faqears/faqears/services/user-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByID(ctx context.Context, id string) (*domain.User, error)
	UpdateProfile(ctx context.Context, id string, p domain.Profile) (*domain.User, error)
}

type FollowRepository interface {
	Add(ctx context.Context, f *domain.Follow) error
	Remove(ctx context.Context, followerID, followeeID string) error
	ListFollowers(ctx context.Context, userID string, limit, offset int) ([]*domain.User, error)
	ListFollowing(ctx context.Context, userID string, limit, offset int) ([]*domain.User, error)
}

type EventPublisher interface {
	ProfileUpdated(ctx context.Context, userID, displayName string) error
	UserFollowed(ctx context.Context, followerID, followeeID string) error
}
