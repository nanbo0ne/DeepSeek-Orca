package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchModels(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/models" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"data": []map[string]string{
				{"id": "model-b", "object": "model"},
				{"id": "model-a", "object": "model"},
			},
		})
	}))
	defer srv.Close()

	models, err := FetchModels(context.Background(), srv.URL, "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 2 {
		t.Fatalf("want 2 models, got %d", len(models))
	}
	if models[0] != "model-a" || models[1] != "model-b" {
		t.Errorf("want sorted [model-a model-b], got %v", models)
	}
}

func TestFetchModelsAuthError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"invalid key"}}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := FetchModels(context.Background(), srv.URL, "bad-key")
	if err == nil {
		t.Fatal("expected error for bad key")
	}
}

func TestFetchModelsEmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": nil})
	}))
	defer srv.Close()

	models, err := FetchModels(context.Background(), srv.URL, "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(models) != 0 {
		t.Errorf("want empty list, got %v", models)
	}
}

func TestFetchModelMetadataReadsVisionHints(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{
			{"id": "vision", "input_modalities": []string{"text", "image"}},
			{"id": "text", "capabilities": map[string]any{"vision": false}},
			{"id": "unknown"},
		}})
	}))
	defer srv.Close()

	items, err := FetchModelMetadata(context.Background(), srv.URL, "key")
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]ModelMetadata{}
	for _, item := range items {
		byID[item.ID] = item
	}
	if byID["vision"].Vision == nil || !*byID["vision"].Vision {
		t.Fatalf("vision metadata = %+v", byID["vision"])
	}
	if byID["text"].Vision == nil || *byID["text"].Vision {
		t.Fatalf("text metadata = %+v", byID["text"])
	}
	if byID["unknown"].Vision != nil {
		t.Fatalf("unknown metadata = %+v", byID["unknown"])
	}
}

func TestFetchModelMetadataReadsContextCapabilitiesAndPricing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{"data": []map[string]any{
			{
				"id": "capable", "context_length": 131072,
				"supported_parameters": []string{"tools", "structured_outputs"},
				"pricing":              map[string]any{"prompt": "0.000002", "completion": "0.000008", "input_cache_read": "0.0000005", "currency": "$"},
			},
			{"id": "nested", "limits": map[string]any{"context": "65536"}, "capabilities": map[string]any{"tool_calling": false, "structured_output": true}},
		}})
	}))
	defer srv.Close()

	items, err := FetchModelMetadata(context.Background(), srv.URL, "key")
	if err != nil {
		t.Fatal(err)
	}
	byID := map[string]ModelMetadata{}
	for _, item := range items {
		byID[item.ID] = item
	}
	capable := byID["capable"]
	if capable.ContextWindow != 131072 || !capable.ContextWindowConfirmed || capable.ContextReason != "context_length" {
		t.Fatalf("context metadata = %+v", capable)
	}
	if capable.ToolUse == nil || !*capable.ToolUse || capable.StructuredOutput == nil || !*capable.StructuredOutput {
		t.Fatalf("capabilities = %+v", capable)
	}
	if capable.Pricing == nil || capable.Pricing.Input != 2 || capable.Pricing.Output != 8 || capable.Pricing.CacheHit != 0.5 || capable.Pricing.Currency != "$" {
		t.Fatalf("pricing = %+v", capable.Pricing)
	}
	nested := byID["nested"]
	if nested.ContextWindow != 65536 || nested.ToolUse == nil || *nested.ToolUse || nested.StructuredOutput == nil || !*nested.StructuredOutput {
		t.Fatalf("nested metadata = %+v", nested)
	}
}
