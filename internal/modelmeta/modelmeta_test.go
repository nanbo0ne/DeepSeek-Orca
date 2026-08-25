package modelmeta

import (
	"path/filepath"
	"testing"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/provider"
)

func testEntry() *config.ProviderEntry {
	return &config.ProviderEntry{
		Name: "custom", Kind: "openai", BaseURL: "https://example.com/v1", Model: "model-a",
		ContextWindow: 32000, ModelContextWindows: map[string]int{"model-a": 64000},
	}
}

func TestResolveContextPriority(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metadata.json")
	store := Load(path)
	entry := testEntry()
	if got := Resolve(entry, store); got.ContextWindow != 64000 || !got.ContextConfirmed || got.ContextSource != SourceUserOverride {
		t.Fatalf("user override = %+v", got)
	}
	if err := store.Put(Metadata{Key: Key(entry), ModelRef: ModelRef(entry), ContextWindow: 128000, ContextConfirmed: true, ContextSource: SourceProviderMetadata}); err != nil {
		t.Fatal(err)
	}
	if got := Resolve(entry, Load(path)); got.ContextWindow != 128000 || got.ContextSource != SourceProviderMetadata {
		t.Fatalf("discovered context = %+v", got)
	}
	local := *entry
	local.Name = "orca-local"
	local.ContextWindow = 8192
	if got := Resolve(&local, store); got.ContextWindow != 8192 || got.ContextSource != SourceLocalRuntime {
		t.Fatalf("local runtime context = %+v", got)
	}
}

func TestSparseRefreshPreservesConfirmedMetadata(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metadata.json")
	store := Load(path)
	entry := testEntry()
	first := MetadataFromDiscovery(entry, 128000, "context_length", CapabilitySupported, CapabilitySupported, CapabilitySupported, &provider.Pricing{Input: 2, Currency: "$"})
	if err := store.Put(first); err != nil {
		t.Fatal(err)
	}
	if err := Load(path).Put(MetadataFromDiscovery(entry, 0, "", CapabilityUnknown, CapabilityUnknown, CapabilityUnknown, nil)); err != nil {
		t.Fatal(err)
	}
	got := Load(path).Get(entry)
	if got.ContextWindow != 128000 || got.Vision != CapabilitySupported || got.ToolUse != CapabilitySupported || got.Pricing == nil {
		t.Fatalf("sparse refresh erased metadata: %+v", got)
	}
}

func TestParallelStoresMerge(t *testing.T) {
	path := filepath.Join(t.TempDir(), "metadata.json")
	a := Load(path)
	b := Load(path)
	if err := a.Put(Metadata{Key: "a", ContextWindow: 1}); err != nil {
		t.Fatal(err)
	}
	if err := b.Put(Metadata{Key: "b", ContextWindow: 2}); err != nil {
		t.Fatal(err)
	}
	if got := Load(path); len(got.Items) != 2 {
		t.Fatalf("items = %+v", got.Items)
	}
}

func TestOfficialDeepSeekCapabilities(t *testing.T) {
	store := Load(filepath.Join(t.TempDir(), "metadata.json"))
	base := config.ProviderEntry{Name: "deepseek", Kind: "openai", BaseURL: "https://api.deepseek.com"}
	tests := []struct {
		model, vision string
	}{
		{"deepseek-v4-flash", CapabilityUnsupported},
		{"deepseek-v4-pro", CapabilityUnsupported},
		{"deepseek-v4-flash-vision-exp", CapabilitySupported},
	}
	for _, tt := range tests {
		entry := base
		entry.Model = tt.model
		got := Resolve(&entry, store)
		if got.Vision != tt.vision || got.ToolUse != CapabilitySupported || got.StructuredOutput != CapabilitySupported {
			t.Fatalf("%s capabilities = %+v", tt.model, got)
		}
	}
	relay := base
	relay.Name, relay.BaseURL, relay.Model = "relay", "https://relay.example/v1", "deepseek-v4-flash-vision-exp"
	if got := Resolve(&relay, store); got.Vision != CapabilityUnknown {
		t.Fatalf("relay inherited official capability: %+v", got)
	}
}
