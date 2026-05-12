package domain

import "github.com/faqears/faqears/pkg/errs"

var (
	ErrNoRecommendations = errs.NotFound("no recommendations found")
)
