package oauth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/auth-service/internal/port"
)

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
}

type Google struct {
	cfg GoogleConfig
}

func NewGoogle(cfg GoogleConfig) *Google {
	if cfg.AuthURL == "" {
		cfg.AuthURL = "https://accounts.google.com/o/oauth2/v2/auth"
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = "https://oauth2.googleapis.com/token"
	}
	return &Google{cfg: cfg}
}

func (g *Google) Name() string {
	return "google"
}

func (g *Google) AuthCodeURL(state string, redirectURI string) string {
	u, _ := url.Parse(g.cfg.AuthURL)
	q := u.Query()
	q.Set("client_id", g.cfg.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "openid email profile")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return fmt.Sprintf("%s", u.String())
}

func (g *Google) Exchange(ctx context.Context, code string, redirectURI string) (*port.OAuthUserInfo, error) {
	return nil, errs.New(errs.KindUnavailable, "google oauth not implemented")
}
