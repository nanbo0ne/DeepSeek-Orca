package main

import "testing"

func TestRuntimeReconfigureOnlyLatestGenerationCanFinish(t *testing.T) {
	tab := &WorkspaceTab{ID: "tab-1"}
	app := &App{tabs: map[string]*WorkspaceTab{tab.ID: tab}}

	first := app.beginTabRuntimeReconfigure(tab)
	latest := app.beginTabRuntimeReconfigure(tab)
	if first == latest {
		t.Fatal("runtime generations must be unique")
	}

	app.finishTabRuntimeReconfigure(tab, first, true)
	if !tab.runtimeReconfiguring {
		t.Fatal("a stale build must not clear the latest reconfiguration state")
	}
	if !app.tabRuntimeGenerationCurrent(tab, latest) {
		t.Fatal("latest generation should remain current")
	}

	app.finishTabRuntimeReconfigure(tab, latest, false)
	if tab.runtimeReconfiguring {
		t.Fatal("the latest failed build should settle the reconfiguration state")
	}
}

func TestFailedRuntimeReconfigureRetainsQueuedSubmissions(t *testing.T) {
	tab := &WorkspaceTab{
		ID: "tab-1",
		pendingRuntimeSubmits: []pendingRuntimeSubmit{{
			input: "keep me queued",
		}},
	}
	app := &App{tabs: map[string]*WorkspaceTab{tab.ID: tab}}

	generation := app.beginTabRuntimeReconfigure(tab)
	app.finishTabRuntimeReconfigure(tab, generation, false)

	if got := len(tab.pendingRuntimeSubmits); got != 1 {
		t.Fatalf("failed switch dropped queued submissions: got %d", got)
	}
}
