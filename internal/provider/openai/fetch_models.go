package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"
)

type modelFetchStatusError struct {
	status int
	body   string
}

func (e modelFetchStatusError) Error() string {
	return fmt.Sprintf("fetch models: status %d: %s", e.status, strings.TrimSpace(e.body))
}

// IsModelFetchEndpointMiss reports whether a model-list request reached a
// plausible endpoint path that the provider does not implement.
func IsModelFetchEndpointMiss(err error) bool {
	var statusErr modelFetchStatusError
	if !errors.As(err, &statusErr) {
		return false
	}
	return statusErr.status == http.StatusNotFound || statusErr.status == http.StatusMethodNotAllowed
}

type ModelMetadata struct {
	ID           string
	Vision       *bool
	VisionReason string
}

// FetchModels calls the OpenAI-compatible GET /models endpoint and returns the
// available model IDs.
func FetchModels(ctx context.Context, baseURL, apiKey string) ([]string, error) {
	items, err := FetchModelMetadata(ctx, baseURL, apiKey)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids, nil
}

// FetchModelMetadata retains optional modality hints exposed by compatible
// gateways while keeping FetchModels backward compatible for existing callers.
func FetchModelMetadata(ctx context.Context, baseURL, apiKey string) ([]ModelMetadata, error) {
	cli := &http.Client{Timeout: 10 * time.Second}
	url := strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(url, "/models") {
		url += "/models"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch models: build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch models: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	if err != nil {
		return nil, fmt.Errorf("fetch models: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, modelFetchStatusError{status: resp.StatusCode, body: truncateFetchBody(string(body))}
	}

	var result struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("fetch models: decode response: %w", err)
	}

	items := make([]ModelMetadata, 0, len(result.Data))
	for _, m := range result.Data {
		id, _ := m["id"].(string)
		if id == "" {
			continue
		}
		vision, reason := modelVisionMetadata(m)
		items = append(items, ModelMetadata{ID: id, Vision: vision, VisionReason: reason})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].ID < items[j].ID })
	return items, nil
}

func modelVisionMetadata(raw map[string]any) (*bool, string) {
	for _, key := range []string{"input_modalities", "supported_input_modalities", "modalities"} {
		if values, ok := stringList(raw[key]); ok {
			return boolPtr(containsVisionModality(values)), key
		}
	}
	if architecture, ok := raw["architecture"].(map[string]any); ok {
		if values, ok := stringList(architecture["input_modalities"]); ok {
			return boolPtr(containsVisionModality(values)), "architecture.input_modalities"
		}
	}
	if capabilities, ok := raw["capabilities"].(map[string]any); ok {
		for _, key := range []string{"vision", "image", "multimodal"} {
			if value, ok := capabilities[key].(bool); ok {
				return boolPtr(value), "capabilities." + key
			}
		}
	}
	return nil, ""
}

func stringList(value any) ([]string, bool) {
	items, ok := value.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			out = append(out, strings.ToLower(strings.TrimSpace(text)))
		}
	}
	return out, true
}

func containsVisionModality(values []string) bool {
	for _, value := range values {
		if value == "image" || value == "vision" || value == "image_url" || value == "multimodal" {
			return true
		}
	}
	return false
}

func boolPtr(value bool) *bool { return &value }

func truncateFetchBody(body string) string {
	body = strings.TrimSpace(body)
	const max = 512
	if len([]rune(body)) <= max {
		return body
	}
	r := []rune(body)
	return string(r[:max]) + "..."
}
