package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pkgredis "github.com/faqears/faqears/pkg/redis"
	"github.com/faqears/faqears/services/playlist-service/internal/domain"
)

type Cache struct {
	client *pkgredis.Client
}

func NewCache(client *pkgredis.Client) *Cache {
	return &Cache{client: client}
}

func playlistKey(id string) string {
	return fmt.Sprintf("playlist:%s", id)
}

func permalinkKey(permalink string) string {
	return fmt.Sprintf("playlist:permalink:%s", permalink)
}

func trackPlaylistsKey(trackID string) string {
	return fmt.Sprintf("track:playlists:%s", trackID)
}

func (c *Cache) GetPlaylist(ctx context.Context, id string) (*domain.Playlist, error) {
	data, err := c.client.Get(ctx, playlistKey(id))
	if err != nil || data == nil {
		return nil, err
	}
	p := &domain.Playlist{}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (c *Cache) SetPlaylist(ctx context.Context, p *domain.Playlist, ttl time.Duration) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, playlistKey(p.ID), data, ttl)
}

func (c *Cache) GetPlaylistByPermalink(ctx context.Context, permalink string) (*domain.Playlist, error) {
	data, err := c.client.Get(ctx, permalinkKey(permalink))
	if err != nil || data == nil {
		return nil, err
	}
	p := &domain.Playlist{}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (c *Cache) SetPlaylistByPermalink(ctx context.Context, p *domain.Playlist, ttl time.Duration) error {
	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, permalinkKey(p.Permalink), data, ttl)
}

func (c *Cache) InvalidatePlaylist(ctx context.Context, id, permalink string) error {
	return c.client.Del(ctx, playlistKey(id), permalinkKey(permalink))
}

func (c *Cache) AddTrackPlaylistRef(ctx context.Context, trackID, playlistID string) error {
	return c.client.Raw().SAdd(ctx, trackPlaylistsKey(trackID), playlistID).Err()
}

func (c *Cache) RemoveTrackPlaylistRef(ctx context.Context, trackID, playlistID string) error {
	return c.client.Raw().SRem(ctx, trackPlaylistsKey(trackID), playlistID).Err()
}

func (c *Cache) GetPlaylistsForTrack(ctx context.Context, trackID string) ([]string, error) {
	return c.client.Raw().SMembers(ctx, trackPlaylistsKey(trackID)).Result()
}
