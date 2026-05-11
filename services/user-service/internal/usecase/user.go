package usecase

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/logger"
	"github.com/faqears/faqears/services/user-service/internal/domain"
	"github.com/faqears/faqears/services/user-service/internal/port"
)

type User struct {
	users   port.UserRepository
	follows port.FollowRepository
	events  port.EventPublisher
}

func NewUser(users port.UserRepository, follows port.FollowRepository, events port.EventPublisher) *User {
	return &User{users: users, follows: follows, events: events}
}

func (u *User) Provision(ctx context.Context, id, email string) error {
	now := time.Now().UTC()
	existing, err := u.users.GetByID(ctx, id)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return errs.Internal("lookup user", err)
	}
	if existing != nil {
		return nil
	}
	rec := &domain.User{
		ID:        id,
		Email:     email,
		Language:  "en",
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := u.users.Create(ctx, rec); err != nil {
		return errs.Internal("create user", err)
	}
	return nil
}

func (u *User) Get(ctx context.Context, id string) (*domain.User, error) {
	if id == "" {
		return nil, domain.ErrInvalidUserID
	}
	rec, err := u.users.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, errs.Internal("get user", err)
	}
	return rec, nil
}

func (u *User) UpdateProfile(ctx context.Context, id string, p domain.Profile) (*domain.User, error) {
	if id == "" {
		return nil, domain.ErrInvalidUserID
	}
	rec, err := u.users.UpdateProfile(ctx, id, p)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, errs.Internal("update profile", err)
	}
	if err := u.events.ProfileUpdated(ctx, rec.ID, rec.DisplayName); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "publish profile_updated failed",
			slog.String("user_id", rec.ID),
			slog.String("error", err.Error()),
		)
	}
	return rec, nil
}

func (u *User) Follow(ctx context.Context, followerID, followeeID string) error {
	if followerID == "" || followeeID == "" {
		return domain.ErrInvalidUserID
	}
	if followerID == followeeID {
		return domain.ErrSelfFollow
	}
	f := &domain.Follow{
		FollowerID: followerID,
		FolloweeID: followeeID,
		CreatedAt:  time.Now().UTC(),
	}
	if err := u.follows.Add(ctx, f); err != nil {
		return errs.Internal("add follow", err)
	}
	if err := u.events.UserFollowed(ctx, followerID, followeeID); err != nil {
		logger.FromContext(ctx).WarnContext(ctx, "publish user_followed failed",
			slog.String("follower_id", followerID),
			slog.String("followee_id", followeeID),
			slog.String("error", err.Error()),
		)
	}
	return nil
}

func (u *User) Unfollow(ctx context.Context, followerID, followeeID string) error {
	if followerID == "" || followeeID == "" {
		return domain.ErrInvalidUserID
	}
	if err := u.follows.Remove(ctx, followerID, followeeID); err != nil {
		return errs.Internal("remove follow", err)
	}
	return nil
}

func (u *User) ListFollowers(ctx context.Context, id string, limit, offset int) ([]*domain.User, error) {
	if id == "" {
		return nil, domain.ErrInvalidUserID
	}
	return u.follows.ListFollowers(ctx, id, normalizeLimit(limit), normalizeOffset(offset))
}

func (u *User) ListFollowing(ctx context.Context, id string, limit, offset int) ([]*domain.User, error) {
	if id == "" {
		return nil, domain.ErrInvalidUserID
	}
	return u.follows.ListFollowing(ctx, id, normalizeLimit(limit), normalizeOffset(offset))
}

func normalizeLimit(l int) int {
	if l <= 0 || l > 100 {
		return 50
	}
	return l
}

func normalizeOffset(o int) int {
	if o < 0 {
		return 0
	}
	return o
}
