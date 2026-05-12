package port

import (
	"context"

	"github.com/faqears/faqears/services/generation-service/internal/domain"
)

type JobRepository interface {
	Create(ctx context.Context, j *domain.Job) error
	GetByID(ctx context.Context, id string) (*domain.Job, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Job, error)
	UpdateStatus(ctx context.Context, id, status string) error
	UpdateStatusAndLyrics(ctx context.Context, id, status, lyrics string) error
	UpdateStatusAndFailure(ctx context.Context, id, status, reason string) error
	UpdatePublished(ctx context.Context, id, trackID string) error
	UpdateStatusField(ctx context.Context, id, status string) error
}

type LyricsRepository interface {
	Save(ctx context.Context, lv *domain.LyricsVersion) error
	ListByJob(ctx context.Context, jobID string) ([]*domain.LyricsVersion, error)
	LatestRevision(ctx context.Context, jobID string) (int32, error)
}

type LyricsGenerator interface {
	GenerateLyrics(ctx context.Context, prompt, genre, mood, language, title string) (string, error)
	ReviseLyrics(ctx context.Context, currentLyrics, instruction string) (string, error)
}

type MusicResult struct {
	Data        []byte
	ContentType string
}

type MusicGenerator interface {
	Name() string
	IsConfigured() bool
	Generate(ctx context.Context, lyrics, genre, mood string) (*MusicResult, error)
}

type StreamingClient interface {
	UploadAudio(ctx context.Context, trackID, contentType string, data []byte) error
}

type CatalogClient interface {
	IngestTrack(ctx context.Context, title, ownerUserID string, genres []string) (string, error)
}

type UserClient interface {
	IsPremium(ctx context.Context, userID string) (bool, error)
}

type EventPublisher interface {
	Publish(ctx context.Context, eventType, jobID, userID string) error
	Close() error
}
