package grpc

import (
	"context"
	"time"

	recov1 "github.com/faqears/faqears/gen/go/recommendation/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/recommendation-service/internal/domain"
	"github.com/faqears/faqears/services/recommendation-service/internal/usecase"
)

type Server struct {
	recov1.UnimplementedRecommendationServiceServer
	uc *usecase.Recommendation
}

func NewServer(uc *usecase.Recommendation) *Server {
	return &Server{uc: uc}
}

func (s *Server) GetDailyMix(ctx context.Context, req *recov1.GetDailyMixRequest) (*recov1.GetDailyMixResponse, error) {
	refs, err := s.uc.GetDailyMix(ctx, req.GetUserId(), req.GetLimit())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &recov1.GetDailyMixResponse{
		Tracks:      refsToProto(refs),
		GeneratedAt: time.Now().Unix(),
	}, nil
}

func (s *Server) GetDiscoverWeekly(ctx context.Context, req *recov1.GetDiscoverWeeklyRequest) (*recov1.GetDiscoverWeeklyResponse, error) {
	refs, err := s.uc.GetDiscoverWeekly(ctx, req.GetUserId(), req.GetLimit())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &recov1.GetDiscoverWeeklyResponse{
		Tracks:      refsToProto(refs),
		GeneratedAt: time.Now().Unix(),
	}, nil
}

func (s *Server) GetSimilarTracks(ctx context.Context, req *recov1.GetSimilarTracksRequest) (*recov1.GetSimilarTracksResponse, error) {
	refs, err := s.uc.GetSimilarTracks(ctx, req.GetTrackId(), req.GetLimit())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &recov1.GetSimilarTracksResponse{Tracks: refsToProto(refs)}, nil
}

func (s *Server) GetTopTracks(ctx context.Context, req *recov1.GetTopTracksRequest) (*recov1.GetTopTracksResponse, error) {
	refs, err := s.uc.GetTopTracks(ctx, req.GetLimit())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &recov1.GetTopTracksResponse{Tracks: refsToProto(refs)}, nil
}

func refsToProto(refs []*domain.TrackRef) []*recov1.TrackRef {
	out := make([]*recov1.TrackRef, len(refs))
	for i, r := range refs {
		out[i] = &recov1.TrackRef{TrackId: r.TrackID, Score: r.Score}
	}
	return out
}
