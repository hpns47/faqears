package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/pkg/grpcx"
	"github.com/faqears/faqears/services/streaming-service/internal/domain"
	"github.com/faqears/faqears/services/streaming-service/internal/port"
)

type Streaming struct {
	audio     port.AudioRepository
	sessions  port.SessionStore
	storage   port.Storage
	events    port.EventPublisher
	bucket    string
	signedTTL time.Duration
}

func NewStreaming(
	audio port.AudioRepository,
	sessions port.SessionStore,
	storage port.Storage,
	events port.EventPublisher,
	bucket string,
	signedTTL time.Duration,
) *Streaming {
	return &Streaming{
		audio:     audio,
		sessions:  sessions,
		storage:   storage,
		events:    events,
		bucket:    bucket,
		signedTTL: signedTTL,
	}
}

var allowedContentTypes = []string{"audio/mpeg", "audio/mp4", "audio/ogg", "audio/wav"}

func (s *Streaming) UploadAudio(ctx context.Context, trackID, contentType string, data []byte) (*domain.TrackAudio, error) {
	if !grpcx.HasRole(ctx, "admin") && !grpcx.HasRole(ctx, "service") && !grpcx.HasRole(ctx, "user") {
		return nil, errs.PermissionDenied("authentication required")
	}

	if !strings.HasPrefix(contentType, "audio/") {
		return nil, domain.ErrInvalidContentType
	}

	allowed := false
	for _, ct := range allowedContentTypes {
		if contentType == ct {
			allowed = true
			break
		}
	}
	if !allowed {
		return nil, domain.ErrInvalidContentType
	}

	exists, _, err := s.audio.Exists(ctx, trackID)
	if err != nil {
		return nil, errs.Internal("check audio exists", err)
	}
	if exists {
		return nil, domain.ErrAlreadyExists
	}

	ext := extensionFor(contentType)
	objectKey := fmt.Sprintf("tracks/%s/%s%s", trackID, uuid.NewString(), ext)

	size, err := s.storage.Put(ctx, objectKey, contentType, data)
	if err != nil {
		return nil, errs.Internal("upload to storage", err)
	}

	userID := grpcx.UserIDFromCtx(ctx)
	a := &domain.TrackAudio{
		TrackID:     trackID,
		ObjectKey:   objectKey,
		ContentType: contentType,
		SizeBytes:   size,
		UploadedBy:  userID,
		UploadedAt:  time.Now().UTC(),
	}
	if err := s.audio.Upsert(ctx, a); err != nil {
		return nil, errs.Internal("upsert audio record", err)
	}
	return a, nil
}

func (s *Streaming) GetStreamURL(ctx context.Context, trackID string) (string, int64, int64, string, error) {
	if grpcx.UserIDFromCtx(ctx) == "" {
		return "", 0, 0, "", errs.Unauthenticated("missing user identity")
	}
	a, err := s.audio.GetByTrackID(ctx, trackID)
	if err != nil {
		return "", 0, 0, "", errs.Internal("get audio", err)
	}
	if a == nil {
		return "", 0, 0, "", domain.ErrAudioNotFound
	}
	url, err := s.storage.PresignGet(ctx, a.ObjectKey, s.signedTTL)
	if err != nil {
		return "", 0, 0, "", errs.Internal("presign url", err)
	}
	expiresAt := time.Now().Add(s.signedTTL).Unix()
	return url, expiresAt, a.SizeBytes, a.ContentType, nil
}

func (s *Streaming) RegisterPlay(ctx context.Context, trackID, sessionID string) (string, error) {
	userID := grpcx.UserIDFromCtx(ctx)
	if userID == "" {
		return "", errs.Unauthenticated("missing user identity")
	}
	if sessionID == "" {
		sessionID = uuid.NewString()
	}
	sess := &domain.Session{
		SessionID: sessionID,
		UserID:    userID,
		TrackID:   trackID,
		StartedAt: time.Now().UTC().Unix(),
	}
	if err := s.sessions.Save(ctx, sess, time.Hour); err != nil {
		return "", errs.Internal("save session", err)
	}
	if err := s.events.PlayStarted(ctx, userID, trackID, sessionID); err != nil {
		_ = err
	}
	return sessionID, nil
}

func (s *Streaming) RegisterSkip(ctx context.Context, trackID, sessionID string, positionSec int32) error {
	userID := grpcx.UserIDFromCtx(ctx)
	if userID == "" {
		return errs.Unauthenticated("missing user identity")
	}
	if err := s.events.PlaySkipped(ctx, userID, trackID, sessionID, positionSec); err != nil {
		_ = err
	}
	return nil
}

func (s *Streaming) RegisterComplete(ctx context.Context, trackID, sessionID string, durationSec int32) error {
	userID := grpcx.UserIDFromCtx(ctx)
	if userID == "" {
		return errs.Unauthenticated("missing user identity")
	}
	if err := s.events.PlayCompleted(ctx, userID, trackID, sessionID, durationSec); err != nil {
		_ = err
	}
	if err := s.sessions.Delete(ctx, sessionID); err != nil {
		_ = err
	}
	return nil
}

func (s *Streaming) GetPlaybackSession(ctx context.Context, sessionID string) (*domain.Session, error) {
	if grpcx.UserIDFromCtx(ctx) == "" {
		return nil, errs.Unauthenticated("missing user identity")
	}
	sess, err := s.sessions.Get(ctx, sessionID)
	if err != nil {
		return nil, errs.Internal("get session", err)
	}
	if sess == nil {
		return nil, domain.ErrSessionNotFound
	}
	return sess, nil
}

func (s *Streaming) HasAudio(ctx context.Context, trackID string) (bool, int64, error) {
	if grpcx.UserIDFromCtx(ctx) == "" {
		return false, 0, errs.Unauthenticated("missing user identity")
	}
	exists, size, err := s.audio.Exists(ctx, trackID)
	if err != nil {
		return false, 0, errs.Internal("check audio exists", err)
	}
	return exists, size, nil
}

func extensionFor(ct string) string {
	switch ct {
	case "audio/mpeg":
		return ".mp3"
	case "audio/mp4":
		return ".m4a"
	case "audio/ogg":
		return ".ogg"
	case "audio/wav":
		return ".wav"
	default:
		return ""
	}
}
