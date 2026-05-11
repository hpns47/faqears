package jwt

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/faqears/faqears/services/auth-service/internal/domain"
)

type Config struct {
	Secret     string
	Issuer     string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

type Issuer struct {
	cfg Config
}

func New(cfg Config) *Issuer {
	return &Issuer{cfg: cfg}
}

type claims struct {
	Email string   `json:"email"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (i *Issuer) IssuePair(_ context.Context, u *domain.User) (*domain.TokenPair, error) {
	now := time.Now().UTC()
	accessExp := now.Add(i.cfg.AccessTTL)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims{
		Email: u.Email,
		Roles: u.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			Issuer:    i.cfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(accessExp),
		},
	})
	access, err := token.SignedString([]byte(i.cfg.Secret))
	if err != nil {
		return nil, err
	}
	refresh, err := randomToken(32)
	if err != nil {
		return nil, err
	}
	return &domain.TokenPair{
		AccessToken:     access,
		RefreshToken:    refresh,
		AccessExpiresAt: accessExp,
	}, nil
}

func (i *Issuer) Validate(_ context.Context, accessToken string) (*domain.Claims, error) {
	parsed, err := jwt.ParseWithClaims(accessToken, &claims{}, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(i.cfg.Secret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, errors.New("invalid token")
	}
	c, ok := parsed.Claims.(*claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return &domain.Claims{
		UserID: c.Subject,
		Email:  c.Email,
		Roles:  c.Roles,
	}, nil
}

func (i *Issuer) HashRefresh(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
