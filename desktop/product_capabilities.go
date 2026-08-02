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
// three-state PromptMode remains intact for old sessions and the future Orca app.
type ProductCapabilities struct {
	Edition                    string   `json:"edition"`
	PromptModes                []string `json:"promptModes"`
	AssistantMemoryEnabled     bool     `json:"assistantMemoryEnabled"`
	AutomationWorkspaceEnabled bool     `json:"automationWorkspaceEnabled"`
}

func productCapabilities() ProductCapabilities {
	if currentProductEdition == productEditionAssistant {
		return ProductCapabilities{
			Edition:                    productEditionAssistant,
			PromptModes:                []string{promptModeAssistant},
			AssistantMemoryEnabled:     true,
			AutomationWorkspaceEnabled: true,
		}
	}
	return ProductCapabilities{
		Edition:                    productEditionEngineering,
		PromptModes:                []string{promptModeNormal, promptModeEnhanced},
		AssistantMemoryEnabled:     false,
		AutomationWorkspaceEnabled: true,
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
	return promptModeNormal
}

func validateProductPromptMode(mode string) (string, error) {
	normalized := normalizePromptMode(mode, false)
	if !productSupportsPromptMode(normalized) {
		return "", fmt.Errorf("prompt mode %q is not available in the %s edition", normalized, currentProductEdition)
	}
	return normalized, nil
}
