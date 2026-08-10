package main

import "testing"

func TestProductCapabilitiesExposeCodingAssistantAndOrca(t *testing.T) {
	capabilities := productCapabilities()
	if capabilities.Edition != productEditionEngineering || !capabilities.AssistantMemoryEnabled || !capabilities.OrcaEnabled {
		t.Fatalf("capabilities = %+v, want coding/assistant product with Orca", capabilities)
	}
	if len(capabilities.PromptModes) != 2 || capabilities.PromptModes[0] != promptModeAssistant || capabilities.PromptModes[1] != promptModeCoding {
		t.Fatalf("prompt modes = %#v, want assistant/coding", capabilities.PromptModes)
	}
	if got := normalizeProductPromptMode(promptModeEnhanced, true); got != promptModeCoding {
		t.Fatalf("enhanced migration = %q, want coding", got)
	}
	if _, err := validateProductPromptMode(promptModeAssistant); err != nil {
		t.Fatalf("assistant mode rejected: %v", err)
	}
}

func TestAssistantMemoryWorkerSupportsAssistantAndOrca(t *testing.T) {
	if !assistantMemoryAvailable() {
		t.Fatal("assistant memory should be available to Assistant and Orca profiles")
	}
	if !productCapabilities().AutomationWorkspaceEnabled || !assistantMemoryFeatureAvailable(automationWorkspaceRoot()) {
		t.Fatal("automation workspace should retain the internal assistant memory worker")
	}
	app := NewApp()
	app.schedulePendingAssistantMemories()
	if app.assistantMemoryTimer == nil {
		t.Fatal("automation memory worker should create an idle timer")
	}
	app.assistantMemoryTimer.Stop()
}
