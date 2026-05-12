package aigen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/faqears/faqears/pkg/errs"
)

const anthropicSystemPrompt = "You are a songwriter. Write a song with the requested theme, mood, genre, and language. Use sections labeled [Verse 1], [Chorus], [Verse 2], [Bridge]. Keep it under 300 words."

type AnthropicGenerator struct {
	apiKey string
	model  string
	client *http.Client
}

func NewAnthropicGenerator(apiKey, model string) *AnthropicGenerator {
	if model == "" {
		model = "claude-haiku-4-5-20251001"
	}
	return &AnthropicGenerator{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 60 * time.Second},
	}
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	Error   *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (g *AnthropicGenerator) call(ctx context.Context, userPrompt string) (string, error) {
	if g.apiKey == "" {
		return "", errs.New(errs.KindUnavailable, "anthropic api key not configured")
	}
	body := anthropicRequest{
		Model:     g.model,
		MaxTokens: 1024,
		System:    anthropicSystemPrompt,
		Messages:  []anthropicMessage{{Role: "user", Content: userPrompt}},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", errs.Internal("marshal anthropic request", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(payload))
	if err != nil {
		return "", errs.Internal("create anthropic request", err)
	}
	req.Header.Set("x-api-key", g.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", errs.Wrap(errs.KindUnavailable, "anthropic request failed", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var ar anthropicResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return "", errs.Internal("decode anthropic response", err)
	}
	if ar.Error != nil {
		return "", errs.New(errs.KindUnavailable, "anthropic error: "+ar.Error.Message)
	}
	if len(ar.Content) == 0 {
		return "", errs.New(errs.KindInternal, "anthropic returned empty content")
	}
	return ar.Content[0].Text, nil
}

func (g *AnthropicGenerator) GenerateLyrics(ctx context.Context, prompt, genre, mood, language, title string) (string, error) {
	userPrompt := fmt.Sprintf("Title: %s\nTheme/Prompt: %s\nGenre: %s\nMood: %s\nLanguage: %s\n\nPlease write the song now.", title, prompt, genre, mood, language)
	return g.call(ctx, userPrompt)
}

func (g *AnthropicGenerator) ReviseLyrics(ctx context.Context, currentLyrics, instruction string) (string, error) {
	userPrompt := fmt.Sprintf("Here are the current lyrics:\n\n%s\n\nRevision instruction: %s\n\nPlease rewrite the lyrics incorporating the instruction.", currentLyrics, instruction)
	return g.call(ctx, userPrompt)
}
