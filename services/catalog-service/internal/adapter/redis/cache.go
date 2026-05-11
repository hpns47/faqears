package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pkgredis "github.com/faqears/faqears/pkg/redis"
	"github.com/faqears/faqears/services/catalog-service/internal/domain"
)

type Cache struct {
	client *pkgredis.Client
}

func NewCache(client *pkgredis.Client) *Cache {
	return &Cache{client: client}
}

func (c *Cache) GetArtist(ctx context.Context, id string) (*domain.Artist, error) {
	data, err := c.client.Get(ctx, fmt.Sprintf("catalog:artist:%s", id))
	if err != nil || data == nil {
		return nil, err
	}
	a := &domain.Artist{}
	if err := json.Unmarshal(data, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (c *Cache) SetArtist(ctx context.Context, a *domain.Artist, ttl time.Duration) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fmt.Sprintf("catalog:artist:%s", a.ID), data, ttl)
}

func (c *Cache) GetAlbum(ctx context.Context, id string) (*domain.Album, error) {
	data, err := c.client.Get(ctx, fmt.Sprintf("catalog:album:%s", id))
	if err != nil || data == nil {
		return nil, err
	}
	a := &domain.Album{}
	if err := json.Unmarshal(data, a); err != nil {
		return nil, err
	}
	return a, nil
}

func (c *Cache) SetAlbum(ctx context.Context, a *domain.Album, ttl time.Duration) error {
	data, err := json.Marshal(a)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fmt.Sprintf("catalog:album:%s", a.ID), data, ttl)
}

func (c *Cache) GetTrack(ctx context.Context, id string) (*domain.Track, error) {
	data, err := c.client.Get(ctx, fmt.Sprintf("catalog:track:%s", id))
	if err != nil || data == nil {
		return nil, err
	}
	t := &domain.Track{}
	if err := json.Unmarshal(data, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (c *Cache) SetTrack(ctx context.Context, t *domain.Track, ttl time.Duration) error {
	data, err := json.Marshal(t)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, fmt.Sprintf("catalog:track:%s", t.ID), data, ttl)
}
