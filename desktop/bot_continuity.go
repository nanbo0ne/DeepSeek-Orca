package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"deepseek-orca/internal/agent"
	"deepseek-orca/internal/boot"
	"deepseek-orca/internal/bot"
	"deepseek-orca/internal/config"
	"deepseek-orca/internal/control"
	"deepseek-orca/internal/provider"
)

func (a *App) decideBotContinuity(ctx context.Context, modelRef string, previous bot.SessionChoice, current string) (bool, error) {
	cfg, err := config.LoadForRoot(previous.WorkspaceRoot)
	if err != nil {
		return false, err
	}
	resolved, _, ok := cfg.ResolveModelWithFallback(modelRef)
	if !ok {
		return false, fmt.Errorf("unknown bot model %q", modelRef)
	}
	entry, ok := cfg.ResolveModel(resolved)
	if !ok {
		return false, fmt.Errorf("unknown bot model %q", resolved)
	}
	prov, err := boot.NewProviderWithProxy(entry, cfg.NetworkProxySpec())
	if err != nil {
		return false, err
	}
	userTail, assistantTail := continuitySessionTail(previous.Path)
	prompt := fmt.Sprintf("Previous segment title: %s\nRecent user message: %s\nPrevious final answer: %s\nCurrent message: %s\n\nReply only RELATED or NEW.", previous.Title, userTail, assistantTail, strings.TrimSpace(current))
	decisionCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	stream, err := prov.Stream(decisionCtx, provider.Request{Temperature: 0, MaxTokens: 12, Messages: []provider.Message{
		{Role: provider.RoleSystem, Content: "Decide whether a new mobile assistant message continues the prior conversation segment. RELATED only when carrying over prior context is clearly useful; otherwise NEW."},
		{Role: provider.RoleUser, Content: prompt},
	}})
	if err != nil {
		return false, err
	}
	var out strings.Builder
	for chunk := range stream {
		if chunk.Type == provider.ChunkError && chunk.Err != nil {
			return false, chunk.Err
		}
		if chunk.Type == provider.ChunkText {
			out.WriteString(chunk.Text)
		}
	}
	answer := strings.ToUpper(strings.TrimSpace(out.String()))
	if strings.HasPrefix(answer, "RELATED") {
		return true, nil
	}
	if strings.HasPrefix(answer, "NEW") {
		return false, nil
	}
	return false, fmt.Errorf("invalid continuity decision %q", answer)
}

func continuitySessionTail(path string) (string, string) {
	session, err := agent.LoadSession(path)
	if err != nil {
		return "", ""
	}
	var user, assistant string
	for _, message := range session.Snapshot() {
		content := strings.TrimSpace(message.Content)
		switch message.Role {
		case provider.RoleUser:
			content = control.StripComposePrefixes(content)
			if content != "" && !control.IsSyntheticUserMessage(content) {
				user = brokerOneLine(content, 600)
			}
		case provider.RoleAssistant:
			if content != "" {
				assistant = brokerOneLine(content, 800)
			}
		}
	}
	return user, assistant
}
