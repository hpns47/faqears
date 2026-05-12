package aigen

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/faqears/faqears/pkg/errs"
	"github.com/faqears/faqears/services/generation-service/internal/port"
)

const replicateBaseURL = "https://api.replicate.com/v1"
const replicateModel = "meta/musicgen"

type ReplicateGenerator struct {
	apiToken string
	client   *http.Client
}

func NewReplicateGenerator(apiToken string) *ReplicateGenerator {
	return &ReplicateGenerator{
		apiToken: apiToken,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (r *ReplicateGenerator) Name() string { return "replicate" }

func (r *ReplicateGenerator) IsConfigured() bool { return r.apiToken != "" }

type replicatePredictionRequest struct {
	Version string                 `json:"version"`
	Input   map[string]interface{} `json:"input"`
}

type replicatePredictionResponse struct {
	ID     string   `json:"id"`
	Status string   `json:"status"`
	Output []string `json:"output"`
	Error  string   `json:"error"`
	URLs   struct {
		Get string `json:"get"`
	} `json:"urls"`
}

func (r *ReplicateGenerator) Generate(ctx context.Context, lyrics, genre, mood string) (*port.MusicResult, error) {
	if !r.IsConfigured() {
		return nil, errs.New(errs.KindUnavailable, "replicate api token not configured")
	}

	predReq := replicatePredictionRequest{
		Version: "b05b1dff1d8c6dc63d14b0cdb42135378dcb87f6373b0d3d341ede46e59e2b38",
		Input: map[string]interface{}{
			"prompt":            lyrics + " genre:" + genre + " mood:" + mood,
			"model_version":     "stereo-large",
			"output_format":     "mp3",
			"normalization_strategy": "peak",
		},
	}
	payload, err := json.Marshal(predReq)
	if err != nil {
		return nil, errs.Internal("marshal replicate request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, replicateBaseURL+"/predictions", bytes.NewReader(payload))
	if err != nil {
		return nil, errs.Internal("create replicate request", err)
	}
	req.Header.Set("Authorization", "Token "+r.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		slog.WarnContext(ctx, "replicate create prediction failed", slog.String("error", err.Error()))
		return nil, errs.Wrap(errs.KindUnavailable, "replicate create prediction failed", err)
	}
	defer resp.Body.Close()

	var predResp replicatePredictionResponse
	if err := json.NewDecoder(resp.Body).Decode(&predResp); err != nil {
		return nil, errs.Internal("decode replicate response", err)
	}
	if predResp.ID == "" {
		return nil, errs.New(errs.KindUnavailable, "replicate returned no prediction id")
	}

	pollURL := replicateBaseURL + "/predictions/" + predResp.ID
	for i := 0; i < 60; i++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}

		pollReq, err := http.NewRequestWithContext(ctx, http.MethodGet, pollURL, nil)
		if err != nil {
			return nil, errs.Internal("create replicate poll request", err)
		}
		pollReq.Header.Set("Authorization", "Token "+r.apiToken)

		pollResp, err := r.client.Do(pollReq)
		if err != nil {
			slog.WarnContext(ctx, "replicate poll failed", slog.String("error", err.Error()))
			return nil, errs.Wrap(errs.KindUnavailable, "replicate poll failed", err)
		}

		var sr replicatePredictionResponse
		json.NewDecoder(pollResp.Body).Decode(&sr)
		pollResp.Body.Close()

		if sr.Status == "succeeded" {
			if len(sr.Output) == 0 {
				return nil, errs.New(errs.KindUnavailable, "replicate succeeded but no output")
			}
			dlReq, err := http.NewRequestWithContext(ctx, http.MethodGet, sr.Output[0], nil)
			if err != nil {
				return nil, errs.Internal("create replicate download request", err)
			}
			dlResp, err := r.client.Do(dlReq)
			if err != nil {
				return nil, errs.Wrap(errs.KindUnavailable, "replicate audio download failed", err)
			}
			defer dlResp.Body.Close()
			data, err := io.ReadAll(dlResp.Body)
			if err != nil {
				return nil, errs.Internal("read replicate audio", err)
			}
			return &port.MusicResult{Data: data, ContentType: "audio/mpeg"}, nil
		}
		if sr.Status == "failed" || sr.Status == "canceled" {
			return nil, errs.New(errs.KindUnavailable, "replicate generation failed: "+sr.Error)
		}
	}
	return nil, errs.New(errs.KindUnavailable, "replicate generation timed out")
}
