package main

import (
	"testing"
	"time"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/visioncap"
)

func TestShouldAutoProbeVision(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)
	for _, tc := range []struct {
		name string
		cap  visioncap.Capability
		want bool
	}{
		{name: "fresh unknown", cap: visioncap.Capability{Status: visioncap.Unknown}, want: true},
		{name: "supported is final", cap: visioncap.Capability{Status: visioncap.Supported}, want: false},
		{name: "unsupported is final", cap: visioncap.Capability{Status: visioncap.Unsupported}, want: false},
		{name: "active probe", cap: visioncap.Capability{Status: visioncap.Probing}, want: false},
		{name: "recent transient failure", cap: visioncap.Capability{Status: visioncap.Unknown, Attempts: 1, CheckedAt: now.Add(-23 * time.Hour).UnixMilli()}, want: false},
		{name: "old transient failure", cap: visioncap.Capability{Status: visioncap.Unknown, Attempts: 2, CheckedAt: now.Add(-25 * time.Hour).UnixMilli()}, want: true},
		{name: "retry limit", cap: visioncap.Capability{Status: visioncap.Unknown, Attempts: 3, CheckedAt: now.Add(-48 * time.Hour).UnixMilli()}, want: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldAutoProbeVision(tc.cap, now); got != tc.want {
				t.Fatalf("shouldAutoProbeVision(%+v) = %v, want %v", tc.cap, got, tc.want)
			}
		})
	}
}
