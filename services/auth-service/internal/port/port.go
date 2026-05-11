package port

import (
	"context"

	"github.com/faqears/faqears/services/auth-service/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, u *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id string) (*domain.User, error)
}

type RefreshTokenRepository interface {
	Save(ctx context.Context, t *domain.RefreshToken) error
	GetByHash(ctx context.Context, hash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(password, hash string) error
}

type TokenIssuer interface {
	IssuePair(ctx context.Context, u *domain.User) (*domain.TokenPair, error)
	Validate(ctx context.Context, accessToken string) (*domain.Claims, error)
	HashRefresh(token string) string
}

type EventPublisher interface {
	UserRegistered(ctx context.Context, userID, email string) error
	UserLoggedIn(ctx context.Context, userID string) error
}
