package grpc

import (
	"context"

	generationv1 "github.com/faqears/faqears/gen/go/generation/v1"
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/services/generation-service/internal/domain"
	"github.com/faqears/faqears/services/generation-service/internal/usecase"
)

type Server struct {
	generationv1.UnimplementedGenerationServiceServer
	uc *usecase.Generation
}

func NewServer(uc *usecase.Generation) *Server {
	return &Server{uc: uc}
}

func (s *Server) GenerateLyrics(ctx context.Context, req *generationv1.GenerateLyricsRequest) (*generationv1.GenerateLyricsResponse, error) {
	callerID := grpcx.UserIDFromCtx(ctx)
	if callerID == "" {
		return nil, errs.ToGRPC(errs.Unauthenticated("unauthenticated"))
	}
	userID := req.GetUserId()
	if userID == "" {
		userID = callerID
	}
	job, lv, err := s.uc.GenerateLyrics(ctx, userID, req.GetPrompt(), req.GetGenre(), req.GetMood(), req.GetLanguage(), req.GetTitle())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &generationv1.GenerateLyricsResponse{
		Job:    domainJobToProto(job),
		Lyrics: domainLyricsToProto(lv),
	}, nil
}

func (s *Server) ReviseLyrics(ctx context.Context, req *generationv1.ReviseLyricsRequest) (*generationv1.ReviseLyricsResponse, error) {
	callerID := grpcx.UserIDFromCtx(ctx)
	if callerID == "" {
		return nil, errs.ToGRPC(errs.Unauthenticated("unauthenticated"))
	}
	userID := req.GetUserId()
	if userID == "" {
		userID = callerID
	}
	job, lv, err := s.uc.ReviseLyrics(ctx, req.GetJobId(), userID, req.GetInstruction())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &generationv1.ReviseLyricsResponse{
		Job:    domainJobToProto(job),
		Lyrics: domainLyricsToProto(lv),
	}, nil
}

func (s *Server) GenerateMusic(ctx context.Context, req *generationv1.GenerateMusicRequest) (*generationv1.GenerateMusicResponse, error) {
	callerID := grpcx.UserIDFromCtx(ctx)
	if callerID == "" {
		return nil, errs.ToGRPC(errs.Unauthenticated("unauthenticated"))
	}
	userID := req.GetUserId()
	if userID == "" {
		userID = callerID
	}
	job, err := s.uc.GenerateMusic(ctx, req.GetJobId(), userID)
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &generationv1.GenerateMusicResponse{Job: domainJobToProto(job)}, nil
}

func (s *Server) GetJob(ctx context.Context, req *generationv1.GetJobRequest) (*generationv1.GetJobResponse, error) {
	callerID := grpcx.UserIDFromCtx(ctx)
	if callerID == "" {
		return nil, errs.ToGRPC(errs.Unauthenticated("unauthenticated"))
	}
	job, history, err := s.uc.GetJob(ctx, req.GetJobId())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	resp := &generationv1.GetJobResponse{Job: domainJobToProto(job)}
	for _, lv := range history {
		resp.LyricsHistory = append(resp.LyricsHistory, domainLyricsToProto(lv))
	}
	return resp, nil
}

func (s *Server) ListUserJobs(ctx context.Context, req *generationv1.ListUserJobsRequest) (*generationv1.ListUserJobsResponse, error) {
	callerID := grpcx.UserIDFromCtx(ctx)
	if callerID == "" {
		return nil, errs.ToGRPC(errs.Unauthenticated("unauthenticated"))
	}
	userID := req.GetUserId()
	if userID == "" {
		userID = callerID
	}
	jobs, err := s.uc.ListUserJobs(ctx, userID, req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	resp := &generationv1.ListUserJobsResponse{}
	for _, j := range jobs {
		resp.Jobs = append(resp.Jobs, domainJobToProto(j))
	}
	return resp, nil
}

func (s *Server) CancelJob(ctx context.Context, req *generationv1.CancelJobRequest) (*generationv1.CancelJobResponse, error) {
	callerID := grpcx.UserIDFromCtx(ctx)
	if callerID == "" {
		return nil, errs.ToGRPC(errs.Unauthenticated("unauthenticated"))
	}
	userID := req.GetUserId()
	if userID == "" {
		userID = callerID
	}
	job, err := s.uc.CancelJob(ctx, req.GetJobId(), userID)
	if err != nil {
		return nil, errs.ToGRPC(err)
	}
	return &generationv1.CancelJobResponse{Job: domainJobToProto(job)}, nil
}

func domainJobToProto(j *domain.Job) *generationv1.Job {
	if j == nil {
		return nil
	}
	return &generationv1.Job{
		Id:            j.ID,
		UserId:        j.UserID,
		Status:        j.Status,
		Prompt:        j.Prompt,
		Lyrics:        j.Lyrics,
		Title:         j.Title,
		Genre:         j.Genre,
		Mood:          j.Mood,
		Language:      j.Language,
		TrackId:       j.TrackID,
		FailureReason: j.FailureReason,
		CreatedAt:     j.CreatedAt.Unix(),
		UpdatedAt:     j.UpdatedAt.Unix(),
	}
}

func domainLyricsToProto(lv *domain.LyricsVersion) *generationv1.LyricsVersion {
	if lv == nil {
		return nil
	}
	return &generationv1.LyricsVersion{
		Id:        lv.ID,
		JobId:     lv.JobID,
		Revision:  lv.Revision,
		Body:      lv.Body,
		CreatedAt: lv.CreatedAt.Unix(),
	}
}
