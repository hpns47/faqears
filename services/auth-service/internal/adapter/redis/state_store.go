package redis

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"

	pkgredis "github.com/faqears/faqears/pkg/redis"
)

const stateKeyPrefix = "auth:oauth_state:"

type StateStore struct {
	client *pkgredis.Client
}

func NewStateStore(client *pkgredis.Client) *StateStore {
	return &StateStore{client: client}
}

func (s *StateStore) Issue(ctx context.Context, ttl time.Duration) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b)
	key := stateKeyPrefix + token
	if err := s.client.Set(ctx, key, []byte("1"), ttl); err != nil {
		return "", err
	}
	return token, nil
}

func (s *StateStore) Consume(ctx context.Context, state string) (bool, error) {
	key := stateKeyPrefix + state
	val, err := s.client.Get(ctx, key)
	if err != nil {
		return false, err
	}
	if val == nil {
		return false, nil
	}
	if err := s.client.Del(ctx, key); err != nil {
		return false, err
	}
	return true, nil
}
