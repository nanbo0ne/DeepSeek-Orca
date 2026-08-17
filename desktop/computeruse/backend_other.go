//go:build !windows

package computeruse

func NewPlatformBackend() Backend { return nil }
