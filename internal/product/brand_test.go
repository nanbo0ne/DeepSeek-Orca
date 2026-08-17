package product

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCurrentProductSurfacesDoNotUseLegacyBrand(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", ".."))
	paths := []string{
		"README.md",
		"README.en.md",
		"README.DESKTOP.md",
		"README.CLI.md",
		filepath.Join("desktop", "README.md"),
		filepath.Join("desktop", "frontend", "src"),
		filepath.Join("internal", "promptprofile"),
		filepath.Join("npm", "orca-agent"),
		filepath.Join("site", "src"),
		filepath.Join("site", "scripts"),
	}
	legacy := []string{"deepseek-orca", "deepseek reasonix", "deepseek_orca", "deepseekorca"}
	for _, rel := range paths {
		path := filepath.Join(root, rel)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("stat %s: %v", rel, err)
		}
		check := func(name string) {
			if filepath.Base(name) == "attachmentDisplay.ts" {
				return // V2 attachment URL compatibility.
			}
			body, readErr := os.ReadFile(name)
			if readErr != nil {
				t.Errorf("read %s: %v", name, readErr)
				return
			}
			lower := strings.ToLower(string(body))
			for _, forbidden := range legacy {
				if strings.Contains(lower, forbidden) {
					t.Errorf("current product surface %s contains legacy brand %q", name, forbidden)
				}
			}
		}
		if !info.IsDir() {
			check(path)
			continue
		}
		err = filepath.WalkDir(path, func(name string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() && entry.Name() == "__tests__" {
				return fs.SkipDir
			}
			if entry.IsDir() || strings.HasSuffix(entry.Name(), ".map") {
				return nil
			}
			check(name)
			return nil
		})
		if err != nil {
			t.Errorf("scan %s: %v", rel, err)
		}
	}
}
