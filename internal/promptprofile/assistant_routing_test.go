package promptprofile

import (
	"strings"
	"testing"
)

func TestOrcaPromptContainsConversationRoutingContract(t *testing.T) {
	prompt := OrcaSystemPrompt("", "", "", "", "")
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

func TestAssistantPromptDoesNotExposeConversationDispatch(t *testing.T) {
	prompt := AssistantSystemPrompt("", "", "", "", "")
	for _, forbidden := range []string{"conversation_dispatch", "conversation_wait", "Never dispatch to the current automation conversation"} {
		if strings.Contains(prompt, forbidden) {
			t.Fatalf("assistant prompt unexpectedly exposes Orca routing %q", forbidden)
		}
	}
}
