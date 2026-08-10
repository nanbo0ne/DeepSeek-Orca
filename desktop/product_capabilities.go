package main

import (
	"fmt"
	"strings"
)

const (
	productEditionEngineering = "engineering"
	productEditionAssistant   = "assistant"

	currentProductEdition = productEditionEngineering
)

// ProductCapabilities defines the public surface of a product edition. The
// Legacy prompt modes remain readable while the public product exposes only
// Coding and Assistant; Orca is an internal fixed-conversation profile.
type ProductCapabilities struct {
	Edition                    string   `json:"edition"`
	PromptModes                []string `json:"promptModes"`
	AssistantMemoryEnabled     bool     `json:"assistantMemoryEnabled"`
	AutomationWorkspaceEnabled bool     `json:"automationWorkspaceEnabled"`
	ConversationModes          []string `json:"conversationModes"`
	OrcaEnabled                bool     `json:"orcaEnabled"`
}

func productCapabilities() ProductCapabilities {
	if currentProductEdition == productEditionAssistant {
		return ProductCapabilities{
			Edition:                    productEditionAssistant,
			PromptModes:                []string{promptModeAssistant},
			ConversationModes:          []string{promptModeAssistant},
			AssistantMemoryEnabled:     true,
			AutomationWorkspaceEnabled: true,
			OrcaEnabled:                true,
		}
	}
	return ProductCapabilities{
		Edition:                    productEditionEngineering,
		PromptModes:                []string{promptModeCoding, promptModeAssistant},
		ConversationModes:          []string{promptModeCoding, promptModeAssistant},
		AssistantMemoryEnabled:     true,
		AutomationWorkspaceEnabled: true,
		OrcaEnabled:                true,
	}
}

func (a *App) GetProductCapabilities() ProductCapabilities {
	return productCapabilities()
}

func assistantMemoryAvailable() bool {
	return productCapabilities().AssistantMemoryEnabled
}

func productSupportsPromptMode(mode string) bool {
	mode = strings.ToLower(strings.TrimSpace(mode))
	for _, allowed := range productCapabilities().PromptModes {
		if mode == allowed {
			return true
		}
	}
	return false
}

func normalizeProductPromptMode(mode string, enhanced bool) string {
	normalized := normalizePromptMode(mode, enhanced)
	if productSupportsPromptMode(normalized) {
		return normalized
	}
	if currentProductEdition == productEditionAssistant {
		return promptModeAssistant
	}
	return promptModeCoding
}

func validateProductPromptMode(mode string) (string, error) {
	normalized := normalizePromptMode(mode, false)
	if !productSupportsPromptMode(normalized) {
		return "", fmt.Errorf("prompt mode %q is not available in the %s edition", normalized, currentProductEdition)
	}
	return normalized, nil
}
