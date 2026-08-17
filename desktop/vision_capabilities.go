package main

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/boot"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/config"
	"github.com/nanbo0ne/O.R.C.A-for-Windows/internal/visioncap"
)

type VisionCapability = visioncap.Capability

//go:embed build/appicon.png
var visionProbeIconPNG []byte

func (a *App) GetVisionCapabilities() []VisionCapability {
	cfg, err := config.LoadForRoot(a.activeWorkspaceRoot())
	if err != nil {
		return nil
	}
	items := visioncap.Load("").List(cfg)
	a.visionProbeMu.Lock()
	defer a.visionProbeMu.Unlock()
	for i := range items {
		if a.visionProbing[items[i].Key] {
			items[i].AutomaticStatus = visioncap.Probing
			if items[i].Override == "" || items[i].Override == visioncap.OverrideAuto {
				items[i].Status = visioncap.Probing
			}
			items[i].Reason = ""
		}
	}
	return items
}

func (a *App) ProbeModelVision(modelRef string) (VisionCapability, error) {
	modelRef = strings.TrimSpace(modelRef)
	cfg, err := config.LoadForRoot(a.activeWorkspaceRoot())
	if err != nil {
		return VisionCapability{}, err
	}
	entry, ok := cfg.ResolveModel(modelRef)
	if !ok {
		return VisionCapability{}, fmt.Errorf("unknown model %q", modelRef)
	}
	store := visioncap.Load("")
	key := visioncap.Key(entry)
	a.visionProbeMu.Lock()
	if a.visionProbing == nil {
		a.visionProbing = map[string]bool{}
	}
	if a.visionProbing[key] {
		a.visionProbeMu.Unlock()
		return VisionCapability{ModelRef: visioncap.ModelRef(entry), Key: key, Status: visioncap.Probing}, nil
	}
	a.visionProbing[key] = true
	a.visionProbeMu.Unlock()
	defer func() { a.visionProbeMu.Lock(); delete(a.visionProbing, key); a.visionProbeMu.Unlock() }()
	a.visionProbeRunMu.Lock()
	defer a.visionProbeRunMu.Unlock()
	current := store.Stored(entry)
	_ = store.Put(VisionCapability{ModelRef: visioncap.ModelRef(entry), Key: key, Status: visioncap.Probing, Attempts: current.Attempts, Source: visioncap.SourceProbe, Override: current.Override, ProbeVersion: visioncap.CurrentProbeVersion})
	finishUnknown := func(reason error) (VisionCapability, error) {
		result := VisionCapability{
			ModelRef:     visioncap.ModelRef(entry),
			Key:          key,
			Status:       visioncap.Unknown,
			Reason:       "vision probe failed",
			Attempts:     current.Attempts + 1,
			CheckedAt:    time.Now().UnixMilli(),
			Source:       visioncap.SourceProbe,
			Override:     current.Override,
			ProbeVersion: visioncap.CurrentProbeVersion,
		}
		if reason != nil {
			result.Reason = reason.Error()
		}
		if putErr := store.Put(result); putErr != nil {
			return result, putErr
		}
		return result, reason
	}
	prov, err := boot.NewProviderWithProxy(entry, cfg.NetworkProxySpec())
	if err != nil {
		return finishUnknown(err)
	}
	ctx, cancel := context.WithTimeout(a.bootContext(), 45*time.Second)
	defer cancel()
	result := visioncap.Probe(ctx, prov, entry, visionProbeIconPNG)
	result.Override = current.Override
	result.Attempts = current.Attempts + 1
	if result.Status != visioncap.Unknown {
		result.Attempts = 0
	}
	if err := store.Put(result); err != nil {
		return result, err
	}
	if cfg.DesktopVisionMode() == config.VisionModeAuto && current.Status != result.Status && a.activeVisionCapabilityKey(cfg) == key {
		a.refreshVisionRoutingWhenIdle()
	}
	if result.Override == visioncap.Supported || result.Override == visioncap.Unsupported {
		result.Status = result.Override
		result.Source = visioncap.SourceManual
	}
	return result, nil
}

// SetVisionCapabilityOverride records a per-model routing override without
// changing the global off|auto|on vision mode. "auto" clears the override.
func (a *App) SetVisionCapabilityOverride(modelRef, override string) (VisionCapability, error) {
	modelRef = strings.TrimSpace(modelRef)
	override = strings.ToLower(strings.TrimSpace(override))
	if override != visioncap.OverrideAuto && override != visioncap.Supported && override != visioncap.Unsupported {
		return VisionCapability{}, fmt.Errorf("vision override %q is invalid", override)
	}
	cfg, err := config.LoadForRoot(a.activeWorkspaceRoot())
	if err != nil {
		return VisionCapability{}, err
	}
	entry, ok := cfg.ResolveModel(modelRef)
	if !ok {
		return VisionCapability{}, fmt.Errorf("unknown model %q", modelRef)
	}
	store := visioncap.Load("")
	stored := store.Stored(entry)
	stored.Override = override
	if err := store.Put(stored); err != nil {
		return stored, err
	}
	effective := stored
	if override == visioncap.Supported || override == visioncap.Unsupported {
		effective.Status = override
		effective.Source = visioncap.SourceManual
	}
	return effective, nil
}

func (a *App) activeVisionCapabilityKey(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	tab := a.activeTab()
	modelRef := cfg.DefaultModel
	if tab != nil && strings.TrimSpace(tab.model) != "" {
		modelRef = tab.model
	}
	entry, ok := cfg.ResolveModel(modelRef)
	if !ok {
		return ""
	}
	return visioncap.Key(entry)
}

func (a *App) refreshVisionRoutingWhenIdle() {
	tab := a.activeTab()
	if tab == nil || tab.Ctrl == nil || !tab.Ctrl.Running() {
		_ = a.rebuild()
		return
	}
	ctrl := tab.Ctrl
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				if ctrl.Running() {
					continue
				}
				a.mu.RLock()
				stillActive := a.activeTabLocked() == tab && tab.Ctrl == ctrl
				a.mu.RUnlock()
				if stillActive {
					_ = a.rebuild()
				}
				return
			case <-a.bootContext().Done():
				return
			}
		}
	}()
}

func (a *App) scheduleVisionProbes(modelRefs []string) {
	for _, ref := range modelRefs {
		ref := strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		go func() {
			cfg, err := config.LoadForRoot(a.activeWorkspaceRoot())
			if err != nil {
				return
			}
			e, ok := cfg.ResolveModel(ref)
			if !ok || !e.Configured() {
				return
			}
			c := visioncap.Load("").Get(e)
			if !shouldAutoProbeVision(c, time.Now()) {
				return
			}
			_, _ = a.ProbeModelVision(ref)
		}()
	}
}

func (a *App) scheduleVisionProbesForKeyEnv(apiKeyEnv string) {
	apiKeyEnv = strings.TrimSpace(apiKeyEnv)
	if apiKeyEnv == "" {
		return
	}
	cfg, err := config.LoadForRoot(a.activeWorkspaceRoot())
	if err != nil {
		return
	}
	access := providerAccessSet(cfg.Desktop.ProviderAccess)
	refs := make([]string, 0)
	for i := range cfg.Providers {
		entry := &cfg.Providers[i]
		if entry.APIKeyEnv != apiKeyEnv || !modelProviderAccessAllowed(access, entry.Name) {
			continue
		}
		for _, model := range entry.ChatModelList() {
			refs = append(refs, entry.Name+"/"+model)
		}
	}
	a.scheduleVisionProbes(refs)
}

func shouldAutoProbeVision(c visioncap.Capability, now time.Time) bool {
	if c.Status == visioncap.Supported || c.Status == visioncap.Unsupported || c.Status == visioncap.Probing {
		return false
	}
	if c.Attempts >= 3 {
		return false
	}
	return c.CheckedAt <= 0 || now.Sub(time.UnixMilli(c.CheckedAt)) >= 24*time.Hour
}

func (a *App) scheduleConfiguredVisionProbes() {
	go func() {
		select {
		case <-time.After(10 * time.Second):
		case <-a.bootContext().Done():
			return
		}
		cfg, err := config.LoadForRoot(a.activeWorkspaceRoot())
		if err != nil || cfg.DesktopVisionMode() != config.VisionModeAuto {
			return
		}
		refs := make([]string, 0)
		access := providerAccessSet(cfg.Desktop.ProviderAccess)
		for i := range cfg.Providers {
			providerEntry := &cfg.Providers[i]
			if !providerEntry.Configured() || !modelProviderAccessAllowed(access, providerEntry.Name) {
				continue
			}
			for _, model := range providerEntry.ChatModelList() {
				entry := *providerEntry
				entry.Model = model
				refs = append(refs, visioncap.ModelRef(&entry))
			}
		}
		a.scheduleVisionProbes(refs)
	}()
}
