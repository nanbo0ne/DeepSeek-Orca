package main

import (
	"context"
	"path/filepath"
	"testing"
	"time"
)

func TestSessionExecutionGateSerializesCanonicalPath(t *testing.T) {
	gate := newSessionExecutionGate()
	path := filepath.Join(t.TempDir(), "session.jsonl")
	release, err := gate.Acquire(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	acquired := make(chan struct{})
	go func() {
		secondRelease, secondErr := gate.Acquire(context.Background(), filepath.Join(filepath.Dir(path), ".", filepath.Base(path)))
		if secondErr == nil {
			close(acquired)
			secondRelease()
		}
	}()
	select {
	case <-acquired:
		t.Fatal("second turn acquired the same session before release")
	case <-time.After(50 * time.Millisecond):
	}
	release()
	select {
	case <-acquired:
	case <-time.After(time.Second):
		t.Fatal("second turn did not acquire after release")
	}
}

func TestSessionExecutionGateHonorsCancellation(t *testing.T) {
	gate := newSessionExecutionGate()
	path := filepath.Join(t.TempDir(), "session.jsonl")
	release, err := gate.Acquire(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := gate.Acquire(ctx, path); err == nil {
		t.Fatal("cancelled waiter should fail")
	}
	release()
}
