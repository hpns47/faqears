package minio

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
	Region    string
}

type Client struct {
	c   *minio.Client
	cfg Config
}

func New(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		return nil, errors.New("minio endpoint is empty")
	}
	c, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if _, err := c.ListBuckets(pingCtx); err != nil {
		return nil, fmt.Errorf("ping minio: %w", err)
	}
	return &Client{c: c, cfg: cfg}, nil
}

func (c *Client) Raw() *minio.Client {
	return c.c
}

func (c *Client) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := c.c.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket %s: %w", bucket, err)
	}
	if exists {
		return nil
	}
	if err := c.c.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: c.cfg.Region}); err != nil {
		return fmt.Errorf("create bucket %s: %w", bucket, err)
	}
	return nil
}

func (c *Client) EnsureBuckets(ctx context.Context, buckets ...string) error {
	for _, b := range buckets {
		if err := c.EnsureBucket(ctx, b); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) PutObject(ctx context.Context, bucket, key, contentType string, data []byte) (int64, error) {
	info, err := c.c.PutObject(ctx, bucket, key, bytesReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

func (c *Client) StatObject(ctx context.Context, bucket, key string) (int64, string, error) {
	info, err := c.c.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return 0, "", err
	}
	return info.Size, info.ContentType, nil
}

func (c *Client) PresignGet(ctx context.Context, bucket, key string, ttl time.Duration) (string, error) {
	u, err := c.c.PresignedGetObject(ctx, bucket, key, ttl, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (c *Client) GetObject(ctx context.Context, bucket, key string) ([]byte, string, error) {
	obj, err := c.c.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", err
	}
	defer obj.Close()
	info, err := obj.Stat()
	if err != nil {
		return nil, "", err
	}
	buf := make([]byte, info.Size)
	_, err = obj.Read(buf)
	if err != nil && err.Error() != "EOF" {
		return nil, "", err
	}
	return buf, info.ContentType, nil
}

func bytesReader(b []byte) *byteReader {
	return &byteReader{data: b}
}

type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, errEOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

var errEOF = errors.New("EOF")
