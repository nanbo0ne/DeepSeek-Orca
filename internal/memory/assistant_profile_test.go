package memory

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureCanonicalAssistantStoreImportsLegacyOnce(t *testing.T) {
	root := t.TempDir()
	legacyA := Store{Dir: filepath.Join(root, "projects", "a", "assistant-memory")}
	legacyB := Store{Dir: filepath.Join(root, "projects", "b", "assistant-memory")}
	if _, err := legacyA.Save(Memory{Name: "tone", Title: "回复风格", Description: "偏好简洁", Confidence: 0.8, UpdatedAt: "2026-01-01T00:00:00Z", Body: "old"}); err != nil {
		t.Fatal(err)
	}
	if _, err := legacyB.Save(Memory{Name: "tone-new", Title: "回复风格", Description: "偏好简洁", Confidence: 0.9, UpdatedAt: "2026-02-01T00:00:00Z", Body: "new"}); err != nil {
		t.Fatal(err)
	}

	canonical, err := EnsureCanonicalAssistantStore(root)
	if err != nil {
		t.Fatal(err)
	}
	items := canonical.List()
	if len(items) != 1 || items[0].Body != "new" || items[0].Name != "tone" {
		t.Fatalf("canonical memories = %+v, want newer duplicate under stable name", items)
	}
	if _, err := os.Stat(filepath.Join(root, "assistant-profile", assistantProfileMigrationMarker)); err != nil {
		t.Fatalf("migration marker: %v", err)
	}
	if len(legacyA.List()) != 1 || len(legacyB.List()) != 1 {
		t.Fatal("legacy stores must remain untouched")
	}

	if _, err := legacyA.Save(Memory{Name: "late", Title: "late", Description: "late", Confidence: 1, UpdatedAt: "2026-03-01T00:00:00Z", Body: "must not import"}); err != nil {
		t.Fatal(err)
	}
	canonical, err = EnsureCanonicalAssistantStore(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(canonical.List()) != 1 {
		t.Fatalf("second import ignored marker: %+v", canonical.List())
	}
}
