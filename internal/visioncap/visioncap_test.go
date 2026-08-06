package visioncap

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"deepseek-orca/internal/config"
	"deepseek-orca/internal/provider"
)

type fakeProvider struct {
	chunks []provider.Chunk
	err    error
	req    provider.Request
}

func (p *fakeProvider) Name() string { return "fake" }

func (p *fakeProvider) Stream(_ context.Context, req provider.Request) (<-chan provider.Chunk, error) {
	p.req = req
	if p.err != nil {
		return nil, p.err
	}
	ch := make(chan provider.Chunk, len(p.chunks))
	for _, chunk := range p.chunks {
		ch <- chunk
	}
	close(ch)
	return ch, nil
}

func testEntry() *config.ProviderEntry {
	return &config.ProviderEntry{Name: "Vision", Kind: "OpenAI", BaseURL: "HTTPS://EXAMPLE.COM/v1/", Model: "vision-1"}
}

func TestProbeWithImageSupported(t *testing.T) {
	p := &fakeProvider{chunks: []provider.Chunk{{Type: provider.ChunkText, Text: "4821"}, {Type: provider.ChunkDone}}}
	got := probeWithImage(context.Background(), p, testEntry(), "4821", "TEST_IMAGE_DATA")
	if got.Status != Supported || got.Reason != "" {
		t.Fatalf("capability = %+v, want supported", got)
	}
	if len(p.req.Messages) != 1 || len(p.req.Messages[0].Images) != 1 {
		t.Fatalf("probe request images = %+v", p.req.Messages)
	}
	if p.req.Messages[0].Images[0].Data != "TEST_IMAGE_DATA" {
		t.Fatalf("probe image data = %q", p.req.Messages[0].Images[0].Data)
	}
	if len(p.req.Tools) != 0 || p.req.Temperature != 0 || p.req.MaxTokens != 32 {
		t.Fatalf("probe request = %+v", p.req)
	}
}

func TestProbeAcceptsWrappedStandaloneCode(t *testing.T) {
	for _, answer := range []string{"The digits are 4821.", "`4821`", "结果：4821"} {
		p := &fakeProvider{chunks: []provider.Chunk{{Type: provider.ChunkText, Text: answer}, {Type: provider.ChunkDone}}}
		got := probeWithImage(context.Background(), p, testEntry(), "4821", "data")
		if got.Status != Supported {
			t.Fatalf("answer %q capability = %+v, want supported", answer, got)
		}
	}
	if probeAnswerMatches("148210", "4821") {
		t.Fatal("code embedded in a longer number must not match")
	}
}

func TestProbeWithImageUnsupported(t *testing.T) {
	for _, answer := range []string{"", "I cannot inspect images"} {
		p := &fakeProvider{chunks: []provider.Chunk{{Type: provider.ChunkText, Text: answer}, {Type: provider.ChunkDone}}}
		got := probeWithImage(context.Background(), p, testEntry(), "4821", "data")
		if got.Status != Unsupported || got.Reason == "" {
			t.Fatalf("answer %q capability = %+v, want unsupported with reason", answer, got)
		}
	}
}

func TestProbeWithImageTransportFailureStaysUnknown(t *testing.T) {
	for _, p := range []*fakeProvider{
		{err: errors.New("network unavailable")},
		{chunks: []provider.Chunk{{Type: provider.ChunkError, Err: errors.New("rate limited")}}},
	} {
		got := probeWithImage(context.Background(), p, testEntry(), "4821", "data")
		if got.Status != Unknown || got.Reason == "" {
			t.Fatalf("capability = %+v, want unknown with reason", got)
		}
	}
}

func TestProbeWithImageExplicitRejectionIsUnsupported(t *testing.T) {
	for _, message := range []string{
		"400: image input is not supported by this text-only model",
		"404: No endpoints found that support image input",
	} {
		p := &fakeProvider{err: errors.New(message)}
		got := probeWithImage(context.Background(), p, testEntry(), "4821", "data")
		if got.Status != Unsupported {
			t.Fatalf("capability = %+v, want unsupported", got)
		}
	}
}

func TestManualOverrideMasksStoredProbeResult(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vision.json")
	store := Load(path)
	e := testEntry()
	if err := store.Put(Capability{Key: Key(e), Status: Unsupported, Source: SourceProbe, Override: Supported}); err != nil {
		t.Fatal(err)
	}
	if got := Load(path).Get(e); got.Status != Supported || got.Source != SourceManual {
		t.Fatalf("effective capability = %+v, want manual supported", got)
	}
	if got := Load(path).Stored(e); got.Status != Unsupported || got.Source != SourceProbe {
		t.Fatalf("stored capability = %+v, want probe unsupported", got)
	}
}

func TestCapabilityKeyAndStoreRoundTrip(t *testing.T) {
	e := testEntry()
	if got, want := Key(e), "openai|https://example.com/v1|vision-1"; got != want {
		t.Fatalf("Key() = %q, want %q", got, want)
	}
	path := filepath.Join(t.TempDir(), "vision.json")
	store := Load(path)
	want := Capability{ModelRef: ModelRef(e), Key: Key(e), Status: Supported, CheckedAt: 1234}
	if err := store.Put(want); err != nil {
		t.Fatal(err)
	}
	got := Load(path).Get(e)
	if got.Status != Supported || got.CheckedAt != 1234 || got.ModelRef != want.ModelRef {
		t.Fatalf("round-trip capability = %+v, want %+v", got, want)
	}
}

func TestParallelStoreInstancesMergeResults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vision.json")
	a := Load(path)
	b := Load(path)
	if err := a.Put(Capability{Key: "a", Status: Supported}); err != nil {
		t.Fatal(err)
	}
	if err := b.Put(Capability{Key: "b", Status: Unsupported}); err != nil {
		t.Fatal(err)
	}
	loaded := Load(path)
	if len(loaded.Items) != 2 || loaded.Items["a"].Status != Supported || loaded.Items["b"].Status != Unsupported {
		t.Fatalf("merged items = %+v", loaded.Items)
	}
}
