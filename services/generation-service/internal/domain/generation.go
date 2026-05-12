package domain

import "time"

const (
	StatusLyricsQueued     = "lyrics_queued"
	StatusLyricsInProgress = "lyrics_in_progress"
	StatusLyricsReady      = "lyrics_ready"
	StatusMusicQueued      = "music_queued"
	StatusMusicInProgress  = "music_in_progress"
	StatusIngesting        = "ingesting"
	StatusPublished        = "published"
	StatusFailed           = "failed"
	StatusCancelled        = "cancelled"
)

type Job struct {
	ID            string
	UserID        string
	Status        string
	Prompt        string
	Title         string
	Genre         string
	Mood          string
	Language      string
	Lyrics        string
	TrackID       string
	FailureReason string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type LyricsVersion struct {
	ID        string
	JobID     string
	Revision  int32
	Body      string
	CreatedAt time.Time
}
