package domain

import "time"

type Artist struct {
	ID        string
	Name      string
	Country   string
	Biography string
	CreatedAt time.Time
}

type Album struct {
	ID        string
	ArtistID  string
	Title     string
	Year      int32
	CoverURL  string
	CreatedAt time.Time
}

type Track struct {
	ID            string
	AlbumID       string
	ArtistID      string
	Title         string
	DurationSec   int32
	ISRC          string
	Genres        []string
	UserGenerated bool
	OwnerUserID   string
	CreatedAt     time.Time
}
