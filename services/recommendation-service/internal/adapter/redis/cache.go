package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pkgredis "github.com/faqears/faqears/pkg/redis"
	"github.com/faqears/faqears/services/recommendation-service/internal/domain"
)

const (
	simTTL     = 24 * time.Hour
	dailyTTL   = 24 * time.Hour
	weeklyTTL  = 7 * 24 * time.Hour
	topTTL     = time.Hour
)

type Cache struct {
	client *pkgredis.Client
}

func NewCache(client *pkgredis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) GetSimilar(ctx context.Context, trackID string) ([]*domain.TrackRef, error) {
	return c.getRefs(ctx, fmt.Sprintf("sim:track:%s", trackID))
}

func (c *Cache) SetSimilar(ctx context.Context, trackID string, refs []*domain.TrackRef) error {
	return c.setRefs(ctx, fmt.Sprintf("sim:track:%s", trackID), refs, simTTL)
}

func (c *Cache) GetDaily(ctx context.Context, userID string) ([]*domain.TrackRef, error) {
	return c.getRefs(ctx, fmt.Sprintf("daily:%s", userID))
}

func (c *Cache) SetDaily(ctx context.Context, userID string, refs []*domain.TrackRef) error {
	return c.setRefs(ctx, fmt.Sprintf("daily:%s", userID), refs, dailyTTL)
}

func (c *Cache) GetDiscoverWeekly(ctx context.Context, userID string) ([]*domain.TrackRef, error) {
	return c.getRefs(ctx, fmt.Sprintf("discover_weekly:%s", userID))
}

func (c *Cache) SetDiscoverWeekly(ctx context.Context, userID string, refs []*domain.TrackRef) error {
	return c.setRefs(ctx, fmt.Sprintf("discover_weekly:%s", userID), refs, weeklyTTL)
}

func (c *Cache) GetTopGlobal(ctx context.Context) ([]*domain.TrackRef, error) {
	return c.getRefs(ctx, "top:global")
}

func (c *Cache) SetTopGlobal(ctx context.Context, refs []*domain.TrackRef) error {
	return c.setRefs(ctx, "top:global", refs, topTTL)
}

func (c *Cache) getRefs(ctx context.Context, key string) ([]*domain.TrackRef, error) {
	data, err := c.client.Get(ctx, key)
	if err != nil || data == nil {
		return nil, err
	}
	var refs []*domain.TrackRef
	if err := json.Unmarshal(data, &refs); err != nil {
		return nil, err
	}
	return refs, nil
}

func (c *Cache) setRefs(ctx context.Context, key string, refs []*domain.TrackRef, ttl time.Duration) error {
	data, err := json.Marshal(refs)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttl)
}
