package aigen

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/generation-service/internal/port"
)

type SunoGenerator struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewSunoGenerator(apiKey, baseURL string) *SunoGenerator {
	if baseURL == "" {
		baseURL = "https://api.suno.ai"
	}
	return &SunoGenerator{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *SunoGenerator) Name() string { return "suno" }

func (s *SunoGenerator) IsConfigured() bool { return s.apiKey != "" }

type sunoGenerateRequest struct {
	Prompt string `json:"prompt"`
	Genre  string `json:"genre,omitempty"`
	Mood   string `json:"mood,omitempty"`
}

type sunoGenerateResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

type sunoStatusResponse struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	AudioURL  string `json:"audio_url"`
}

func (s *SunoGenerator) Generate(ctx context.Context, lyrics, genre, mood string) (*port.MusicResult, error) {
	if !s.IsConfigured() {
		return nil, errs.New(errs.KindUnavailable, "suno api key not configured")
	}

	body := sunoGenerateRequest{Prompt: lyrics, Genre: genre, Mood: mood}
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, errs.Internal("marshal suno request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/api/v1/generate", bytes.NewReader(payload))
	if err != nil {
		return nil, errs.Internal("create suno request", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		slog.WarnContext(ctx, "suno generate failed", slog.String("error", err.Error()))
		return nil, errs.Wrap(errs.KindUnavailable, "suno generate failed", err)
	}
	defer resp.Body.Close()

	var genResp sunoGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&genResp); err != nil {
		return nil, errs.Internal("decode suno response", err)
	}
	if genResp.ID == "" {
		return nil, errs.New(errs.KindUnavailable, "suno returned no job id")
	}

	for i := 0; i < 30; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}

		statusReq, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/v1/generate/%s", s.baseURL, genResp.ID), nil)
		if err != nil {
			return nil, errs.Internal("create suno status request", err)
		}
		statusReq.Header.Set("Authorization", "Bearer "+s.apiKey)

		statusResp, err := s.client.Do(statusReq)
		if err != nil {
			slog.WarnContext(ctx, "suno status poll failed", slog.String("error", err.Error()))
			return nil, errs.Wrap(errs.KindUnavailable, "suno status poll failed", err)
		}

		var sr sunoStatusResponse
		json.NewDecoder(statusResp.Body).Decode(&sr)
		statusResp.Body.Close()

		if sr.Status == "complete" || sr.Status == "succeeded" {
			if sr.AudioURL == "" {
				return nil, errs.New(errs.KindUnavailable, "suno completed but no audio url")
			}
			dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, sr.AudioURL, nil)
			if err != nil {
				return nil, errs.Internal("create suno download request", err)
			}
			dlResp, err := s.client.Do(dlReq)
			if err != nil {
				return nil, errs.Wrap(errs.KindUnavailable, "suno audio download failed", err)
			}
			defer dlResp.Body.Close()
			data, err := io.ReadAll(dlResp.Body)
			if err != nil {
				return nil, errs.Internal("read suno audio", err)
			}
			return &port.MusicResult{Data: data, ContentType: "audio/mpeg"}, nil
		}
		if sr.Status == "failed" || sr.Status == "error" {
			return nil, errs.New(errs.KindUnavailable, "suno generation failed")
		}
	}
	return nil, errs.New(errs.KindUnavailable, "suno generation timed out")
}
