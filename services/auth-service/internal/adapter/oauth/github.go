package oauth

import (
	"context"
	"net/url"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/auth-service/internal/port"
)

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	AuthURL      string
	TokenURL     string
}

type GitHub struct {
	cfg GitHubConfig
}

func NewGitHub(cfg GitHubConfig) *GitHub {
	if cfg.AuthURL == "" {
		cfg.AuthURL = "https://github.com/login/oauth/authorize"
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = "https://github.com/login/oauth/access_token"
	}
	return &GitHub{cfg: cfg}
}

func (g *GitHub) Name() string {
	return "github"
}

func (g *GitHub) AuthCodeURL(state string, redirectURI string) string {
	u, _ := url.Parse(g.cfg.AuthURL)
	q := u.Query()
	q.Set("client_id", g.cfg.ClientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", "read:user user:email")
	q.Set("state", state)
	u.RawQuery = q.Encode()
	return u.String()
}

func (g *GitHub) Exchange(ctx context.Context, code string, redirectURI string) (*port.OAuthUserInfo, error) {
	return nil, errs.New(errs.KindUnavailable, "github oauth not implemented")
}
