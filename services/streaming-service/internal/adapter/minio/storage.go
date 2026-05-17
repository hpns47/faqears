package minio

import (
	"context"
	"time"

	pkgminio "github.com/faqears/faqears/pkg/minio"
)

type Storage struct {
	client  *pkgminio.Client
	presign *pkgminio.Client
	bucket  string
}

func NewStorage(client, presign *pkgminio.Client, bucket string) *Storage {
	return &Storage{client: client, presign: presign, bucket: bucket}
}

func (s *Storage) Put(ctx context.Context, key, contentType string, data []byte) (int64, error) {
	return s.client.PutObject(ctx, s.bucket, key, contentType, data)
}

func (s *Storage) PresignGet(ctx context.Context, key string, ttl time.Duration) (string, error) {
	return s.presign.PresignGet(ctx, s.bucket, key, ttl)
}
