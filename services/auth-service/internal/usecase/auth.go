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

func (a *Auth) Login(ctx context.Context, email, password, userAgent string) (*domain.TokenPair, error) {
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
		UserAgent: userAgent,
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

func (a *Auth) Refresh(ctx context.Context, refreshToken, userAgent string) (*domain.TokenPair, error) {
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
	if userAgent == "" {
		userAgent = stored.UserAgent
	}
	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    u.ID,
		TokenHash: a.issuer.HashRefresh(pair.RefreshToken),
		ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
		UserAgent: userAgent,
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

func (a *Auth) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	u, err := a.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return errs.Internal("lookup user", err)
	}
	if err := a.hasher.Verify(oldPassword, u.PasswordHash); err != nil {
		return domain.ErrInvalidCredentials
	}
	if err := validatePassword(newPassword); err != nil {
		return err
	}
	if a.hasher.Verify(newPassword, u.PasswordHash) == nil {
		return domain.ErrSamePassword
	}
	hash, err := a.hasher.Hash(newPassword)
	if err != nil {
		return errs.Internal("hash password", err)
	}
	if err := a.users.UpdatePassword(ctx, userID, hash); err != nil {
		return errs.Internal("update password", err)
	}
	return nil
}

func (a *Auth) ListSessions(ctx context.Context, userID, currentRefreshToken string) ([]*domain.Session, error) {
	stored, err := a.tokens.ListByUser(ctx, userID)
	if err != nil {
		return nil, errs.Internal("list sessions", err)
	}
	currentHash := ""
	if currentRefreshToken != "" {
		currentHash = a.issuer.HashRefresh(currentRefreshToken)
	}
	now := time.Now().UTC()
	sessions := make([]*domain.Session, 0, len(stored))
	for _, t := range stored {
		if t.Revoked || t.ExpiresAt.Before(now) {
			continue
		}
		sessions = append(sessions, &domain.Session{
			ID:        t.ID,
			CreatedAt: t.CreatedAt,
			ExpiresAt: t.ExpiresAt,
			Revoked:   t.Revoked,
			Current:   currentHash != "" && t.TokenHash == currentHash,
			UserAgent: t.UserAgent,
		})
	}
	return sessions, nil
}

func (a *Auth) RevokeSession(ctx context.Context, userID, sessionID string) error {
	stored, err := a.tokens.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrSessionNotFound
		}
		return errs.Internal("lookup session", err)
	}
	if stored.UserID != userID {
		return domain.ErrSessionNotFound
	}
	if err := a.tokens.Revoke(ctx, sessionID); err != nil {
		return errs.Internal("revoke session", err)
	}
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
