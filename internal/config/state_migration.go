package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/product"
)

const v11MigrationMarker = ".migration-v11.json"

type v11MigrationRecord struct {
	Version    int       `json:"version"`
	MigratedAt time.Time `json:"migratedAt"`
	LegacyRoot string    `json:"legacyRoot"`
	NewRoot    string    `json:"newRoot"`
}

// EnsureV11StateMigration copies the V2 user tree into the V3 identity once.
// Existing V3 files always win and the legacy tree is retained as the backup.
func EnsureV11StateMigration() error {
	dir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	legacyRoot := filepath.Join(dir, product.LegacyConfigDirName)
	newRoot := filepath.Join(dir, product.ConfigDirName)
	if _, err := os.Stat(legacyRoot); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(newRoot, v11MigrationMarker)); err == nil {
		return nil
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	lockPath := filepath.Join(dir, ".orca-v11-migration.lock")
	lock, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	_ = lock.Close()
	defer os.Remove(lockPath)

	if err := copyTreeMissing(legacyRoot, newRoot); err != nil {
		return fmt.Errorf("migrate V2 state: %w", err)
	}
	record := v11MigrationRecord{Version: 11, MigratedAt: time.Now().UTC(), LegacyRoot: legacyRoot, NewRoot: newRoot}
	body, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return err
	}
	return atomicWriteMigrationFile(filepath.Join(newRoot, v11MigrationMarker), append(body, '\n'), 0o600)
}

func copyTreeMissing(source, target string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return os.MkdirAll(target, 0o755)
		}
		dst := filepath.Join(target, rel)
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if entry.Type()&os.ModeSymlink != 0 {
			// Never follow a legacy symlink outside the owned state tree.
			return nil
		}
		if entry.IsDir() {
			return os.MkdirAll(dst, info.Mode().Perm())
		}
		if _, err := os.Stat(dst); err == nil {
			return nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return copyFileExclusive(path, dst, info.Mode().Perm())
	})
}

func copyFileExclusive(source, target string, mode fs.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(target, os.O_CREATE|os.O_EXCL|os.O_WRONLY, mode)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return nil
		}
		return err
	}
	ok := false
	defer func() {
		_ = out.Close()
		if !ok {
			_ = os.Remove(target)
		}
	}()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := out.Sync(); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	ok = true
	return nil
}

func atomicWriteMigrationFile(path string, body []byte, mode fs.FileMode) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".migration-v11-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(body); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}
