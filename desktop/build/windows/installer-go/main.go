package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

//go:embed payload/*
var payload embed.FS

func main() {
	dir, err := os.MkdirTemp("", "deepcode-setup-*")
	if err != nil {
		fail(err)
	}
	defer os.RemoveAll(dir)

	if err := fs.WalkDir(payload, "payload", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("payload", path)
		if err != nil || rel == "." {
			return err
		}
		target := filepath.Join(dir, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := payload.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o755)
	}); err != nil {
		fail(err)
	}

	cmd := exec.Command(
		"powershell.exe",
		"-ExecutionPolicy", "Bypass",
		"-NoProfile",
		"-STA",
		"-File", filepath.Join(dir, "setup-ui.ps1"),
		"-SourceDir", dir,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err := cmd.Run(); err != nil {
		fail(err)
	}
}

func fail(err error) {
	_ = exec.Command(
		"powershell.exe",
		"-NoProfile",
		"-Command",
		fmt.Sprintf("Add-Type -AssemblyName PresentationFramework; [System.Windows.MessageBox]::Show(%q, 'DeepCode Setup')", err.Error()),
	).Run()
	os.Exit(1)
}
