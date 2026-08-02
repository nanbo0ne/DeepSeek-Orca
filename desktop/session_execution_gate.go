package main

import (
	"context"
	"path/filepath"
	"strings"
	"sync"
)

type sessionExecutionGate struct {
	mu    sync.Mutex
	locks map[string]*sessionExecutionLock
}

type sessionExecutionLock struct {
	token chan struct{}
	refs  int
}

func newSessionExecutionGate() *sessionExecutionGate {
	return &sessionExecutionGate{locks: map[string]*sessionExecutionLock{}}
}

func (g *sessionExecutionGate) Acquire(ctx context.Context, path string) (func(), error) {
	if g == nil {
		return func() {}, nil
	}
	key := canonicalExecutionPath(path)
	if key == "" {
		return func() {}, nil
	}
	g.mu.Lock()
	lock := g.locks[key]
	if lock == nil {
		lock = &sessionExecutionLock{token: make(chan struct{}, 1)}
		lock.token <- struct{}{}
		g.locks[key] = lock
	}
	lock.refs++
	g.mu.Unlock()

	select {
	case <-ctx.Done():
		g.releaseRef(key, lock, false)
		return nil, ctx.Err()
	case <-lock.token:
		var once sync.Once
		return func() { once.Do(func() { g.releaseRef(key, lock, true) }) }, nil
	}
}

func (g *sessionExecutionGate) releaseRef(key string, lock *sessionExecutionLock, held bool) {
	if held {
		lock.token <- struct{}{}
	}
	g.mu.Lock()
	lock.refs--
	if lock.refs == 0 && g.locks[key] == lock {
		delete(g.locks, key)
	}
	g.mu.Unlock()
}

func canonicalExecutionPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	return strings.ToLower(filepath.Clean(path))
}
