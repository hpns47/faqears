package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/services/auth-service/internal/domain"
	"github.com/faqears/faqears/services/auth-service/internal/port"
)

type OAuthResult struct {
	AccessToken     string
	RefreshToken    string
	AccessExpiresAt time.Time
	UserID          string
	Created         bool
}

type OAuthRegistry interface {
	Get(name string) (port.OAuthProvider, error)
}

type OAuth struct {
	registry OAuthRegistry
	store    port.OAuthStateStore
	stateTTL time.Duration
	users    port.UserRepository
	tokens   port.RefreshTokenRepository
	hasher   port.PasswordHasher
	issuer   port.TokenIssuer
	events   port.EventPublisher
}

func NewOAuth(
	registry OAuthRegistry,
	store port.OAuthStateStore,
	stateTTL time.Duration,
	users port.UserRepository,
	tokens port.RefreshTokenRepository,
	hasher port.PasswordHasher,
	issuer port.TokenIssuer,
	events port.EventPublisher,
) *OAuth {
	return &OAuth{
		registry: registry,
		store:    store,
		stateTTL: stateTTL,
		users:    users,
		tokens:   tokens,
		hasher:   hasher,
		issuer:   issuer,
		events:   events,
	}
}

func (o *OAuth) AuthorizeURL(ctx context.Context, provider, redirectURI string) (authorizeURL, state string, err error) {
	p, err := o.registry.Get(provider)
	if err != nil {
		return "", "", err
	}
	state, err = o.store.Issue(ctx, o.stateTTL)
	if err != nil {
		return "", "", err
	}
	return p.AuthCodeURL(state, redirectURI), state, nil
}

func (o *OAuth) Callback(ctx context.Context, provider, code, state string) (*OAuthResult, error) {
	ok, err := o.store.Consume(ctx, state)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errs.New(errs.KindUnauthenticated, "invalid or expired oauth state")
	}
	p, err := o.registry.Get(provider)
	if err != nil {
		return nil, err
	}
	info, err := p.Exchange(ctx, code, "")
	if err != nil {
		return nil, err
	}
	created := false
	u, err := o.users.GetByEmail(ctx, info.Email)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.Internal("lookup user", err)
		}
		randomHash, herr := o.hasher.Hash(uuid.NewString())
		if herr != nil {
			return nil, errs.Internal("generate password hash", herr)
		}
		now := time.Now().UTC()
		u = &domain.User{
			ID:           uuid.NewString(),
			Email:        info.Email,
			PasswordHash: randomHash,
			Roles:        []string{"user"},
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		if cerr := o.users.Create(ctx, u); cerr != nil {
			return nil, errs.Internal("create user", cerr)
		}
		created = true
		if perr := o.events.UserRegistered(ctx, u.ID, u.Email); perr != nil {
			logger.FromContext(ctx).WarnContext(ctx, "publish user_registered failed",
				slog.String("user_id", u.ID),
				slog.String("error", perr.Error()),
			)
		}
	}
	pair, err := o.issuer.IssuePair(ctx, u)
	if err != nil {
		return nil, errs.Internal("issue tokens", err)
	}
	rt := &domain.RefreshToken{
		ID:        uuid.NewString(),
		UserID:    u.ID,
		TokenHash: o.issuer.HashRefresh(pair.RefreshToken),
		ExpiresAt: time.Now().UTC().Add(30 * 24 * time.Hour),
		CreatedAt: time.Now().UTC(),
	}
	if err := o.tokens.Save(ctx, rt); err != nil {
		return nil, errs.Internal("store refresh token", err)
	}
	return &OAuthResult{
		AccessToken:     pair.AccessToken,
		RefreshToken:    pair.RefreshToken,
		AccessExpiresAt: pair.AccessExpiresAt,
		UserID:          u.ID,
		Created:         created,
	}, nil
}
