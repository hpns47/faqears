package aigen

import (
	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/generation-service/internal/port"
)

type Registry struct {
	providers []port.MusicGenerator
}

func NewRegistry(providers ...port.MusicGenerator) *Registry {
	return &Registry{providers: providers}
}

func (r *Registry) Pick() (port.MusicGenerator, error) {
	for _, p := range r.providers {
		if p.IsConfigured() {
			return p, nil
		}
	}
	return nil, errs.New(errs.KindUnavailable, "no music generator configured")
}
