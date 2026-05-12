package grpcclient

import (
	"context"

	catalogv1 "github.com/faqears/faqears/gen/go/catalog/v1"
	"github.com/faqears/faqears/pkg/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type CatalogClient struct {
	client catalogv1.CatalogServiceClient
}

func NewCatalogClient(addr string) (*CatalogClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &CatalogClient{client: catalogv1.NewCatalogServiceClient(conn)}, nil
}

func (c *CatalogClient) IngestTrack(ctx context.Context, title, ownerUserID string, genres []string) (string, error) {
	resp, err := c.client.IngestTrack(ctx, &catalogv1.IngestTrackRequest{
		Title:         title,
		OwnerUserId:   ownerUserID,
		Genres:        genres,
		UserGenerated: true,
	})
	if err != nil {
		return "", errs.Wrap(errs.KindUnavailable, "ingest track failed", err)
	}
	return resp.GetTrackId(), nil
}
