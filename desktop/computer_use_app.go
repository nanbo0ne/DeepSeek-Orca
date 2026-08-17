package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	wruntime "github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/desktop/computeruse"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/localai"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/visioncap"
)

const computerUseConsentVersion = 1

type ComputerUseState struct {
	Capabilities   computeruse.Capabilities `json:"capabilities"`
	Session        computeruse.Session      `json:"session"`
	Approved       bool                     `json:"approved"`
	ConsentVersion int                      `json:"consentVersion"`
	ModelRef       string                   `json:"modelRef,omitempty"`
}

func (a *App) GetComputerUseState() ComputerUseState {
	state := ComputerUseState{}
	if a.computerUse != nil {
		state.Capabilities = a.computerUse.Capabilities()
		state.Session = a.computerUse.Current()
	}
	if cfg, err := config.LoadForRoot(a.activeWorkspaceRoot()); err == nil {
		state.Approved = cfg.Desktop.ComputerUseFullAccess && cfg.Desktop.ComputerUseConsent == computerUseConsentVersion
		state.ConsentVersion = cfg.Desktop.ComputerUseConsent
		state.ModelRef = strings.TrimSpace(cfg.Desktop.ComputerControlModel)
	}
	return state
}

func (a *App) SetComputerUseFullAccess(enabled bool) error {
	if !enabled && a.computerUse != nil {
		_ = a.computerUse.Stop("authorization revoked")
	}
	return a.applyConfigOnly(func(c *config.Config) error {
		c.Desktop.ComputerUseFullAccess = enabled
		if enabled {
			c.Desktop.ComputerUseConsent = computerUseConsentVersion
		} else {
			c.Desktop.ComputerUseConsent = 0
		}
		return nil
	})
}

func (a *App) StartComputerUseSession(request computeruse.StartRequest) (computeruse.Session, error) {
	if a.computerUse == nil || !a.computerUse.Capabilities().Supported {
		return computeruse.Session{}, computeruse.ErrNotSupported
	}
	cfg, err := config.LoadForRoot(a.activeWorkspaceRoot())
	if err != nil {
		return computeruse.Session{}, err
	}
	if !cfg.Desktop.ComputerUseFullAccess || cfg.Desktop.ComputerUseConsent != computerUseConsentVersion {
		return computeruse.Session{}, fmt.Errorf("使用电脑控制前需要在设置中完成一次完全访问授权")
	}
	modelRef, err := a.resolveComputerControlModel(cfg, request.ModelRef)
	if err != nil {
		return computeruse.Session{}, err
	}
	request.ModelRef = modelRef
	return a.computerUse.Start(a.bootContext(), request)
}

func (a *App) ObserveComputerUse() (computeruse.Observation, error) {
	if a.computerUse == nil {
		return computeruse.Observation{}, computeruse.ErrNotSupported
	}
	return a.computerUse.Observe(a.bootContext())
}

func (a *App) ExecuteComputerAction(action computeruse.Action) (computeruse.ActionResult, error) {
	if a.computerUse == nil {
		return computeruse.ActionResult{}, computeruse.ErrNotSupported
	}
	return a.computerUse.Execute(a.bootContext(), action)
}

func (a *App) PauseComputerUse() (computeruse.Session, error) {
	if a.computerUse == nil {
		return computeruse.Session{}, computeruse.ErrNotSupported
	}
	return a.computerUse.Pause()
}

func (a *App) ResumeComputerUse() (computeruse.Session, error) {
	if a.computerUse == nil {
		return computeruse.Session{}, computeruse.ErrNotSupported
	}
	return a.computerUse.Resume(a.bootContext())
}

func (a *App) StopComputerUse() error {
	if a.computerUse == nil {
		return nil
	}
	return a.computerUse.Stop("stopped by user")
}

func (a *App) resolveComputerControlModel(cfg *config.Config, requested string) (string, error) {
	candidates := []string{strings.TrimSpace(requested), strings.TrimSpace(cfg.Desktop.ComputerControlModel), strings.TrimSpace(cfg.Agent.SubagentModel), strings.TrimSpace(cfg.DefaultModel)}
	seen := map[string]bool{}
	for _, ref := range candidates {
		if ref == "" || seen[ref] {
			continue
		}
		seen[ref] = true
		entry, ok := cfg.ResolveModel(ref)
		if !ok {
			continue
		}
		canonical := entry.Name + "/" + entry.Model
		if entry.Name == localai.ProviderID {
			if spec, ok := localai.ModelByID(entry.Model); ok && spec.Vision && spec.ToolUse {
				return canonical, nil
			}
			continue
		}
		capability := visioncap.Load("").Get(entry)
		if capability.Status == visioncap.Supported {
			return canonical, nil
		}
	}
	return "", fmt.Errorf("没有已确认支持图片和结构化动作的电脑控制模型，请先在设置中选择并完成视觉能力检测")
}

func (a *App) onComputerUseEvent(event computeruse.Event) {
	if a.ctx != nil {
		wruntime.EventsEmit(a.ctx, "computer:session", event.Session)
		if event.Observation != nil {
			wruntime.EventsEmit(a.ctx, "computer:observation", event.Observation)
		}
		if event.Action != nil {
			wruntime.EventsEmit(a.ctx, "computer:action", event.Action)
		}
		if event.Kind == "cancelled" || event.Kind == "failed" || event.Kind == "succeeded" {
			wruntime.EventsEmit(a.ctx, "computer:stopped", event.Session)
		}
	}
	a.appendComputerTelemetry(event)
}

func (a *App) appendComputerTelemetry(event computeruse.Event) {
	type record struct {
		At         time.Time                `json:"at"`
		Kind       string                   `json:"kind"`
		SessionID  string                   `json:"sessionId"`
		State      computeruse.SessionState `json:"state"`
		ActionType string                   `json:"actionType,omitempty"`
		Window     string                   `json:"window,omitempty"`
		DurationMS int64                    `json:"durationMs,omitempty"`
		Success    bool                     `json:"success,omitempty"`
		Error      string                   `json:"error,omitempty"`
	}
	r := record{At: time.Now().UTC(), Kind: event.Kind, SessionID: event.Session.ID, State: event.Session.State}
	if event.Action != nil {
		r.ActionType = event.Action.Action.Type
		r.Window = event.Session.CurrentApp
		r.DurationMS = event.Action.DurationMS
		r.Success = event.Action.Success
		r.Error = event.Action.Error
	}
	body, err := json.Marshal(r)
	if err != nil {
		return
	}
	path := filepath.Join(filepath.Dir(config.UserConfigPath()), "telemetry", "computer-use.jsonl")
	if err = os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(body, '\n'))
}
