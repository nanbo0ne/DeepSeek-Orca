package main

import "testing"

func TestEngineeringProductCapabilitiesHideAssistant(t *testing.T) {
	capabilities := productCapabilities()
	if capabilities.Edition != productEditionEngineering || capabilities.AssistantMemoryEnabled {
		t.Fatalf("capabilities = %+v, want engineering with assistant memory disabled", capabilities)
	}
	if len(capabilities.PromptModes) != 2 || capabilities.PromptModes[0] != promptModeNormal || capabilities.PromptModes[1] != promptModeEnhanced {
		t.Fatalf("prompt modes = %#v, want normal/enhanced", capabilities.PromptModes)
	}
	if got := normalizeProductPromptMode(promptModeAssistant, false); got != promptModeNormal {
		t.Fatalf("assistant restore mode = %q, want normal", got)
	}
	if _, err := validateProductPromptMode(promptModeAssistant); err == nil {
		t.Fatal("engineering edition should reject a new assistant-mode request")
	}
}

func TestEngineeringAssistantMemoryWorkerIsDisabled(t *testing.T) {
	if assistantMemoryAvailable() {
		t.Fatal("engineering edition should not run assistant memory worker")
	}
	app := NewApp()
	app.schedulePendingAssistantMemories()
	if app.assistantMemoryTimer != nil {
		t.Fatal("disabled assistant worker should not create a timer")
	}
}
