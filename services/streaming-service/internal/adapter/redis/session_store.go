package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pkgredis "github.com/faqears/faqears/pkg/redis"
	"github.com/faqears/faqears/services/streaming-service/internal/domain"
)

type SessionStore struct {
	client *pkgredis.Client
}

func NewSessionStore(client *pkgredis.Client) *SessionStore {
	return &SessionStore{client: client}
}

func sessionKey(sessionID string) string {
	return fmt.Sprintf("stream:session:%s", sessionID)
}

func (s *SessionStore) Save(ctx context.Context, sess *domain.Session, ttl time.Duration) error {
	data, err := json.Marshal(sess)
	if err != nil {
		return err
	}
	return s.client.Set(ctx, sessionKey(sess.SessionID), data, ttl)
}

func (s *SessionStore) Get(ctx context.Context, sessionID string) (*domain.Session, error) {
	data, err := s.client.Get(ctx, sessionKey(sessionID))
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	sess := &domain.Session{}
	if err := json.Unmarshal(data, sess); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *SessionStore) Delete(ctx context.Context, sessionID string) error {
	return s.client.Del(ctx, sessionKey(sessionID))
}
