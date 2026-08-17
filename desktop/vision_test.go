package main

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/agent"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/provider"
)

func TestMigrateForkSessionCopiesImageAttachments(t *testing.T) {
	sourceRoot := t.TempDir()
	targetRoot := t.TempDir()
	raw, err := base64.StdEncoding.DecodeString(desktopTinyPNG)
	if err != nil {
		t.Fatal(err)
	}
	rel := filepath.ToSlash(filepath.Join(".deepseek-orca", "attachments", "source.png"))
	sourceImage := filepath.Join(sourceRoot, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(sourceImage), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourceImage, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	session := &agent.Session{Messages: []provider.Message{{Role: provider.RoleUser, Content: "look", Images: []provider.ImageContent{{Path: rel, MediaType: "image/png"}}}}}
	sourceSession := filepath.Join(sourceRoot, "sessions", "fork.jsonl")
	if err := session.Save(sourceSession); err != nil {
		t.Fatal(err)
	}
	targetSession, err := migrateForkSessionToWorkspace(sourceSession, sourceRoot, targetRoot)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := agent.LoadSession(targetSession)
	if err != nil {
		t.Fatal(err)
	}
	got := loaded.Messages[0].Images[0].Path
	if got == rel {
		t.Fatal("fork should rewrite the image reference to its copied snapshot")
	}
	if _, err := os.Stat(filepath.Join(targetRoot, filepath.FromSlash(got))); err != nil {
		t.Fatalf("copied image missing: %v", err)
	}
}
