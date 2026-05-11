package usecase

import (
	"context"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/services/auth-service/internal/domain"
	"github.com/faqears/faqears/services/auth-service/internal/port"
)

type Auth struct {
	users    port.UserRepository
	tokens   port.RefreshTokenRepository
	hasher   port.PasswordHasher
	issuer   port.TokenIssuer
	events   port.EventPublisher
}

func NewAuth(
	users port.UserRepository,
	tokens port.RefreshTokenRepository,
	hasher port.PasswordHasher,
	issuer port.TokenIssuer,
	events port.EventPublisher,
) *Auth {
	return &Auth{
		users:  users,
		tokens: tokens,
		hasher: hasher,
		issuer: issuer,
		events: events,
	}
}

func (a *Auth) Register(ctx context.Context, email, password string) (string, error) {
	email = normalizeEmail(email)
	if err := validateEmail(email); err != nil {
		return "", err
	}
	if err := validatePassword(password); err != nil {
		return "", err
	}
	existing, err := a.users.GetByEmail(ctx, email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return "", errs.Internal("lookup user", err)
	}
	if existing != nil {
		return "", domain.ErrEmailTaken
	}
	hash, err := a.hasher.Hash(password)
	if err != nil {
		return "", errs.Internal("hash password", err)
	}
	now := time.Now().UTC()
	u := &domain.User{
		ID:           uuid.NewString(),
		Email:        email,
		PasswordHash: hash,
		Roles:        []string{"user"},
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := a.users.Create(ctx, u); err != nil {
		return "", errs.Internal("create user", err)
	}
	if err := a.events.UserRegistered(ctx, u.ID, u.Email); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "publish user_registered failed",
			slog.String("user_id", u.ID),
			slog.String("error", err.Error()),
		)
	}
	return u.ID, nil
}

func (a *Auth) Login(ctx context.Context, email, password string) (*domain.TokenPair, error) {
	email = normalizeEmail(email)
	u, err := a.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, errs.Internal("lookup user", err)
	}
	if err := a.hasher.Verify(password, u.PasswordHash); err != nil {
		return nil, domain.ErrInvalidCredentials
	}
	pair, err := a.issuer.IssuePair(ctx, u)
	if err != nil {
		return nil, errs.Internal("issue tokens", err)
	}
	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    u.ID,
		TokenHash: a.issuer.HashRefresh(pair.RefreshToken),
		ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	if err := a.tokens.Save(ctx, rt); err != nil {
		return nil, errs.Internal("store refresh token", err)
	}
	if err := a.events.UserLoggedIn(ctx, u.ID); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "publish user_logged_in failed",
			slog.String("user_id", u.ID),
			slog.String("error", err.Error()),
		)
	}
	return pair, nil
}

func (a *Auth) ValidateToken(ctx context.Context, accessToken string) (*domain.Claims, error) {
	claims, err := a.issuer.Validate(ctx, accessToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}

func (a *Auth) Refresh(ctx context.Context, refreshToken string) (*domain.TokenPair, error) {
	hash := a.issuer.HashRefresh(refreshToken)
	stored, err := a.tokens.GetByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, errs.Internal("lookup refresh token", err)
	}
	if stored.Revoked || stored.ExpiresAt.Before(time.Now().UTC()) {
		return nil, domain.ErrInvalidToken
	}
	u, err := a.users.GetByID(ctx, stored.UserID)
	if err != nil {
		return nil, errs.Internal("lookup user", err)
	}
	pair, err := a.issuer.IssuePair(ctx, u)
	if err != nil {
		return nil, errs.Internal("issue tokens", err)
	}
	if err := a.tokens.Revoke(ctx, stored.ID); err != nil {
		return nil, errs.Internal("revoke old token", err)
	}
	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    u.ID,
		TokenHash: a.issuer.HashRefresh(pair.RefreshToken),
		ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	if err := a.tokens.Save(ctx, rt); err != nil {
		return nil, errs.Internal("store refresh token", err)
	}
	return pair, nil
}

func (a *Auth) Logout(ctx context.Context, accessToken string) error {
	claims, err := a.issuer.Validate(ctx, accessToken)
	if err != nil {
		return domain.ErrInvalidToken
	}
	_ = claims
	return nil
}

func normalizeEmail(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func validateEmail(s string) error {
	if _, err := mail.ParseAddress(s); err != nil {
		return domain.ErrInvalidEmail
	}
	return nil
}

func validatePassword(s string) error {
	if len(s) < 8 {
		return domain.ErrWeakPassword
	}
	return nil
}
