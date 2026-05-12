package grpcclient

import (
	"context"

	streamingv1 "github.com/faqears/faqears/gen/go/streaming/v1"
	"github.com/faqears/faqears/pkg/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type StreamingClient struct {
	client streamingv1.StreamingServiceClient
}

func NewStreamingClient(addr string) (*StreamingClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &StreamingClient{client: streamingv1.NewStreamingServiceClient(conn)}, nil
}

func (c *StreamingClient) UploadAudio(ctx context.Context, trackID, contentType string, data []byte) error {
	_, err := c.client.UploadAudio(ctx, &streamingv1.UploadAudioRequest{
		TrackId:     trackID,
		ContentType: contentType,
		Data:        data,
	})
	if err != nil {
		return errs.Wrap(errs.KindUnavailable, "upload audio failed", err)
	}
	return nil
}

func (c *StreamingClient) HasAudio(ctx context.Context, trackID string) (bool, error) {
	resp, err := c.client.HasAudio(ctx, &streamingv1.HasAudioRequest{TrackId: trackID})
	if err != nil {
		return false, errs.Wrap(errs.KindUnavailable, "has audio failed", err)
	}
	return resp.GetExists(), nil
}
