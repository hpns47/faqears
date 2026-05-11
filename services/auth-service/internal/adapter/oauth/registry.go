package oauth

import (
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/auth-service/internal/port"
)

type Registry struct {
	providers map[string]port.OAuthProvider
}

func NewRegistry(providers ...port.OAuthProvider) *Registry {
	m := make(map[string]port.OAuthProvider, len(providers))
	for _, p := range providers {
		m[p.Name()] = p
	}
	return &Registry{providers: m}
}

func (r *Registry) Get(name string) (port.OAuthProvider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, errs.New(errs.KindInvalidArgument, "unknown oauth provider: "+name)
	}
	return p, nil
}
