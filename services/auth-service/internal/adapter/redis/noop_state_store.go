package redis

import (
	"context"
	"time"

	"github.com/faqears/faqears/pkg/errs"
)

type NoopStateStore struct{}

func NewNoopStateStore() *NoopStateStore {
	return &NoopStateStore{}
}

func (n *NoopStateStore) Issue(ctx context.Context, ttl time.Duration) (string, error) {
	return "", errs.New(errs.KindUnavailable, "oauth state store not configured")
}

func (n *NoopStateStore) Consume(ctx context.Context, state string) (bool, error) {
	return false, errs.New(errs.KindUnavailable, "oauth state store not configured")
}
