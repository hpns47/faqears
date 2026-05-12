package aigen

import (
	"context"
	_ "embed"

	"github.com/faqears/faqears/services/generation-service/internal/port"
)

//go:embed placeholder.mp3
var placeholderMP3 []byte

type MockMusicGenerator struct{}

func NewMockMusicGenerator() *MockMusicGenerator {
	return &MockMusicGenerator{}
}

func (m *MockMusicGenerator) Name() string { return "mock" }

func (m *MockMusicGenerator) IsConfigured() bool { return true }

func (m *MockMusicGenerator) Generate(_ context.Context, _, _, _ string) (*port.MusicResult, error) {
	data := make([]byte, len(placeholderMP3))
	copy(data, placeholderMP3)
	return &port.MusicResult{Data: data, ContentType: "audio/mpeg"}, nil
}
