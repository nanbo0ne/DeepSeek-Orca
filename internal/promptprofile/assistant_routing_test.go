package promptprofile

import (
	"strings"
	"testing"
)

func TestAssistantPromptContainsConversationRoutingContract(t *testing.T) {
	prompt := AssistantSystemPrompt("", "", "", "", "")
	wants := []string{
		"decide within this same turn",
		"Do not call conversation tools merely to demonstrate them",
		"Never claim to have inspected a conversation unless",
		"Never dispatch to the current automation conversation or another automation conversation",
		"conversation_wait",
	}
	for _, want := range wants {
		if !strings.Contains(prompt, want) {
			t.Fatalf("assistant prompt missing %q", want)
		}
	}
}
