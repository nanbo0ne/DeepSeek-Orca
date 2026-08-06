//go:build !windows

package main

func readClipboardFilePaths() ([]string, error) { return []string{}, nil }
