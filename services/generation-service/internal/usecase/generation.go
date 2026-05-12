package usecase

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/services/generation-service/internal/adapter/aigen"
	"github.com/faqears/faqears/services/generation-service/internal/domain"
	"github.com/faqears/faqears/services/generation-service/internal/port"
)

type Generation struct {
	jobs      port.JobRepository
	lyrics    port.LyricsRepository
	lyricsGen port.LyricsGenerator
	musicReg  *aigen.Registry
	streaming port.StreamingClient
	catalog   port.CatalogClient
	user      port.UserClient
	events    port.EventPublisher
	log       *slog.Logger
}

func NewGeneration(
	jobs port.JobRepository,
	lyrics port.LyricsRepository,
	lyricsGen port.LyricsGenerator,
	musicReg *aigen.Registry,
	streaming port.StreamingClient,
	catalog port.CatalogClient,
	user port.UserClient,
	events port.EventPublisher,
	log *slog.Logger,
) *Generation {
	return &Generation{
		jobs:      jobs,
		lyrics:    lyrics,
		lyricsGen: lyricsGen,
		musicReg:  musicReg,
		streaming: streaming,
		catalog:   catalog,
		user:      user,
		events:    events,
		log:       log,
	}
}

func (g *Generation) GenerateLyrics(ctx context.Context, userID, prompt, genre, mood, language, title string) (*domain.Job, *domain.LyricsVersion, error) {
	outCtx := grpcx.OutgoingIdentity(ctx, grpcx.Identity{
		UserID: userID,
		Roles:  append(grpcx.RolesFromCtx(ctx), "service"),
	})

	premium, err := g.user.IsPremium(outCtx, userID)
	if err != nil {
		return nil, nil, err
	}
	if !premium {
		return nil, nil, domain.ErrPremiumRequired
	}

	now := time.Now().UTC()
	job := &domain.Job{
		ID:        uuid.NewString(),
		UserID:    userID,
		Status:    domain.StatusLyricsQueued,
		Prompt:    prompt,
		Title:     title,
		Genre:     genre,
		Mood:      mood,
		Language:  language,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := g.jobs.Create(ctx, job); err != nil {
		return nil, nil, errs.Internal("create job", err)
	}

	if err := g.jobs.UpdateStatus(ctx, job.ID, domain.StatusLyricsInProgress); err != nil {
		return nil, nil, errs.Internal("update job status", err)
	}
	job.Status = domain.StatusLyricsInProgress
	_ = g.events.Publish(ctx, "generation.lyrics_in_progress", job.ID, userID)

	body, err := g.lyricsGen.GenerateLyrics(ctx, prompt, genre, mood, language, title)
	if err != nil {
		_ = g.jobs.UpdateStatusAndFailure(ctx, job.ID, domain.StatusFailed, err.Error())
		_ = g.events.Publish(ctx, "generation.failed", job.ID, userID)
		return nil, nil, err
	}

	lv := &domain.LyricsVersion{
		ID:        uuid.NewString(),
		JobID:     job.ID,
		Revision:  1,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	}
	if err := g.lyrics.Save(ctx, lv); err != nil {
		return nil, nil, errs.Internal("save lyrics version", err)
	}

	if err := g.jobs.UpdateStatusAndLyrics(ctx, job.ID, domain.StatusLyricsReady, body); err != nil {
		return nil, nil, errs.Internal("update job lyrics", err)
	}
	job.Status = domain.StatusLyricsReady
	job.Lyrics = body
	_ = g.events.Publish(ctx, "generation.lyrics_ready", job.ID, userID)

	return job, lv, nil
}

func (g *Generation) ReviseLyrics(ctx context.Context, jobID, userID, instruction string) (*domain.Job, *domain.LyricsVersion, error) {
	job, err := g.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, domain.ErrJobNotFound
	}
	if job.UserID != userID {
		return nil, nil, domain.ErrNotOwner
	}

	rev, err := g.lyrics.LatestRevision(ctx, jobID)
	if err != nil {
		return nil, nil, errs.Internal("get latest revision", err)
	}

	body, err := g.lyricsGen.ReviseLyrics(ctx, job.Lyrics, instruction)
	if err != nil {
		return nil, nil, err
	}

	lv := &domain.LyricsVersion{
		ID:        uuid.NewString(),
		JobID:     jobID,
		Revision:  rev + 1,
		Body:      body,
		CreatedAt: time.Now().UTC(),
	}
	if err := g.lyrics.Save(ctx, lv); err != nil {
		return nil, nil, errs.Internal("save revised lyrics", err)
	}

	if err := g.jobs.UpdateStatusAndLyrics(ctx, jobID, domain.StatusLyricsReady, body); err != nil {
		return nil, nil, errs.Internal("update job revised lyrics", err)
	}
	job.Status = domain.StatusLyricsReady
	job.Lyrics = body

	return job, lv, nil
}

func (g *Generation) GenerateMusic(ctx context.Context, jobID, userID string) (*domain.Job, error) {
	job, err := g.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, domain.ErrJobNotFound
	}
	if job.UserID != userID {
		return nil, domain.ErrNotOwner
	}
	if job.Lyrics == "" {
		return nil, domain.ErrNoLyrics
	}

	if err := g.jobs.UpdateStatus(ctx, jobID, domain.StatusMusicQueued); err != nil {
		return nil, errs.Internal("update status music queued", err)
	}
	job.Status = domain.StatusMusicQueued
	_ = g.events.Publish(ctx, "generation.music_queued", jobID, userID)

	if err := g.jobs.UpdateStatus(ctx, jobID, domain.StatusMusicInProgress); err != nil {
		return nil, errs.Internal("update status music in progress", err)
	}
	job.Status = domain.StatusMusicInProgress
	_ = g.events.Publish(ctx, "generation.music_in_progress", jobID, userID)

	provider, err := g.musicReg.Pick()
	if err != nil {
		_ = g.jobs.UpdateStatusAndFailure(ctx, jobID, domain.StatusFailed, err.Error())
		_ = g.events.Publish(ctx, "generation.failed", jobID, userID)
		return nil, err
	}
	g.log.InfoContext(ctx, "using music provider", slog.String("provider", provider.Name()))

	result, err := provider.Generate(ctx, job.Lyrics, job.Genre, job.Mood)
	if err != nil {
		_ = g.jobs.UpdateStatusAndFailure(ctx, jobID, domain.StatusFailed, err.Error())
		_ = g.events.Publish(ctx, "generation.failed", jobID, userID)
		return nil, err
	}

	if err := g.jobs.UpdateStatus(ctx, jobID, domain.StatusIngesting); err != nil {
		return nil, errs.Internal("update status ingesting", err)
	}
	job.Status = domain.StatusIngesting

	outCtx := grpcx.OutgoingIdentity(ctx, grpcx.Identity{
		UserID: userID,
		Roles:  append(grpcx.RolesFromCtx(ctx), "service"),
	})

	genres := []string{}
	if job.Genre != "" {
		genres = append(genres, job.Genre)
	}

	trackID, err := g.catalog.IngestTrack(outCtx, job.Title, userID, genres)
	if err != nil {
		_ = g.jobs.UpdateStatusAndFailure(ctx, jobID, domain.StatusFailed, err.Error())
		_ = g.events.Publish(ctx, "generation.failed", jobID, userID)
		return nil, err
	}

	if err := g.streaming.UploadAudio(outCtx, trackID, result.ContentType, result.Data); err != nil {
		_ = g.jobs.UpdateStatusAndFailure(ctx, jobID, domain.StatusFailed, err.Error())
		_ = g.events.Publish(ctx, "generation.failed", jobID, userID)
		return nil, err
	}

	if err := g.jobs.UpdatePublished(ctx, jobID, trackID); err != nil {
		return nil, errs.Internal("update job published", err)
	}
	job.Status = domain.StatusPublished
	job.TrackID = trackID
	_ = g.events.Publish(ctx, "generation.published", jobID, userID)

	return job, nil
}

func (g *Generation) GetJob(ctx context.Context, jobID string) (*domain.Job, []*domain.LyricsVersion, error) {
	job, err := g.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, nil, domain.ErrJobNotFound
	}
	history, err := g.lyrics.ListByJob(ctx, jobID)
	if err != nil {
		return nil, nil, errs.Internal("list lyrics history", err)
	}
	return job, history, nil
}

func (g *Generation) ListUserJobs(ctx context.Context, userID string, limit, offset int32) ([]*domain.Job, error) {
	if limit <= 0 {
		limit = 20
	}
	jobs, err := g.jobs.ListByUser(ctx, userID, int(limit), int(offset))
	if err != nil {
		return nil, errs.Internal("list user jobs", err)
	}
	return jobs, nil
}

func (g *Generation) CancelJob(ctx context.Context, jobID, userID string) (*domain.Job, error) {
	job, err := g.jobs.GetByID(ctx, jobID)
	if err != nil {
		return nil, domain.ErrJobNotFound
	}
	if job.UserID != userID {
		return nil, domain.ErrNotOwner
	}
	cancellable := map[string]bool{
		domain.StatusLyricsQueued: true,
		domain.StatusMusicQueued:  true,
	}
	if !cancellable[job.Status] {
		return nil, domain.ErrJobNotCancellable
	}
	if err := g.jobs.UpdateStatus(ctx, jobID, domain.StatusCancelled); err != nil {
		return nil, errs.Internal("cancel job", err)
	}
	job.Status = domain.StatusCancelled
	return job, nil
}
