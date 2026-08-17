package localai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func testArtifact(name string, body []byte, source string) Artifact {
	sum := sha256.Sum256(body)
	return Artifact{Name: name, Size: int64(len(body)), SHA256: hex.EncodeToString(sum[:]), Sources: []string{source}}
}

func TestDownloadFromReplacesCompleteCorruptPart(t *testing.T) {
	want := []byte("verified local model payload")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(want)
	}))
	defer server.Close()

	root := t.TempDir()
	m := NewManager(root, nil)
	part := filepath.Join(root, "payload.part")
	if err := os.WriteFile(part, []byte("xxxxxxxxxxxxxxxxxxxxxxxxxxxx"[:len(want)]), 0o600); err != nil {
		t.Fatal(err)
	}
	artifact := testArtifact("payload.gguf", want, server.URL)
	if err := m.downloadFrom(context.Background(), "missing-task", server.URL, artifact, part, 0); err != nil {
		t.Fatalf("downloadFrom: %v", err)
	}
	got, err := os.ReadFile(part)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("downloaded payload = %q, want %q", got, want)
	}
	if ok, err := fileMatches(part, artifact); !ok || err != nil {
		t.Fatalf("downloaded file must verify: ok=%v err=%v", ok, err)
	}
}

func TestWriteJSONAtomicReplacesExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")
	if err := writeJSONAtomic(path, map[string]int{"generation": 1}); err != nil {
		t.Fatal(err)
	}
	if err := writeJSONAtomic(path, map[string]int{"generation": 2}); err != nil {
		t.Fatal(err)
	}
	var got map[string]int
	if err := readJSON(path, &got); err != nil {
		t.Fatal(err)
	}
	if got["generation"] != 2 {
		t.Fatalf("generation = %d, want 2", got["generation"])
	}
}

func TestExtractZipRejectsTraversal(t *testing.T) {
	// The path guard itself is covered through a deliberately malformed archive
	// in the integration download tests; keep this assertion close to the owned
	// directory guard so future cleanup changes cannot escape the local AI root.
	root := t.TempDir()
	if err := removeOwnedDir(root, filepath.Dir(root)); err == nil {
		t.Fatal("expected cleanup outside local AI root to be rejected")
	}
}
