package grpc

import (
	"context"

	streamingv1 "github.com/faqears/faqears/gen/go/streaming/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/streaming-service/internal/usecase"
)

type Server struct {
	streamingv1.UnimplementedStreamingServiceServer
	uc *usecase.Streaming
}

func NewServer(uc *usecase.Streaming) *Server {
	return &Server{uc: uc}
}

func (s *Server) UploadAudio(ctx context.Context, req *streamingv1.UploadAudioRequest) (*streamingv1.UploadAudioResponse, error) {
	a, err := s.uc.UploadAudio(ctx, req.GetTrackId(), req.GetContentType(), req.GetData())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &streamingv1.UploadAudioResponse{
		ObjectKey: a.ObjectKey,
		SizeBytes: a.SizeBytes,
	}, nil
}

func (s *Server) GetStreamURL(ctx context.Context, req *streamingv1.GetStreamURLRequest) (*streamingv1.GetStreamURLResponse, error) {
	url, expiresAt, size, ct, err := s.uc.GetStreamURL(ctx, req.GetTrackId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &streamingv1.GetStreamURLResponse{
		StreamUrl:   url,
		ExpiresAt:   expiresAt,
		SizeBytes:   size,
		ContentType: ct,
	}, nil
}

func (s *Server) RegisterPlay(ctx context.Context, req *streamingv1.RegisterPlayRequest) (*streamingv1.RegisterPlayResponse, error) {
	sessionID, err := s.uc.RegisterPlay(ctx, req.GetTrackId(), req.GetSessionId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &streamingv1.RegisterPlayResponse{SessionId: sessionID}, nil
}

func (s *Server) RegisterSkip(ctx context.Context, req *streamingv1.RegisterSkipRequest) (*streamingv1.RegisterSkipResponse, error) {
	if err := s.uc.RegisterSkip(ctx, req.GetTrackId(), req.GetSessionId(), req.GetPositionSec()); err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &streamingv1.RegisterSkipResponse{}, nil
}

func (s *Server) RegisterComplete(ctx context.Context, req *streamingv1.RegisterCompleteRequest) (*streamingv1.RegisterCompleteResponse, error) {
	if err := s.uc.RegisterComplete(ctx, req.GetTrackId(), req.GetSessionId(), req.GetDurationSec()); err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &streamingv1.RegisterCompleteResponse{}, nil
}

func (s *Server) GetPlaybackSession(ctx context.Context, req *streamingv1.GetPlaybackSessionRequest) (*streamingv1.GetPlaybackSessionResponse, error) {
	sess, err := s.uc.GetPlaybackSession(ctx, req.GetSessionId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &streamingv1.GetPlaybackSessionResponse{
		SessionId: sess.SessionID,
		UserId:    sess.UserID,
		TrackId:   sess.TrackID,
		StartedAt: sess.StartedAt,
	}, nil
}

func (s *Server) HasAudio(ctx context.Context, req *streamingv1.HasAudioRequest) (*streamingv1.HasAudioResponse, error) {
	exists, size, err := s.uc.HasAudio(ctx, req.GetTrackId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &streamingv1.HasAudioResponse{
		Exists:    exists,
		SizeBytes: size,
	}, nil
}
