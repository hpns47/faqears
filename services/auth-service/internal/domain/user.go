package domain

import "time"

type User struct {
	ID           string
	Email        string
	PasswordHash string
	Roles        []string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
	Revoked   bool
	UserAgent string
}

type Session struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
	Revoked   bool
	Current   bool
	UserAgent string
}

type TokenPair struct {
	AccessToken     string
	RefreshToken    string
	AccessExpiresAt time.Time
}

type Claims struct {
	UserID string
	Email  string
	Roles  []string
}
