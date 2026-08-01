package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"deepseek-orca/internal/boot"
	"deepseek-orca/internal/config"
	"deepseek-orca/internal/visioncap"
)

type VisionCapability = visioncap.Capability

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
			items[i].Status = visioncap.Probing
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
	current := store.Get(entry)
	_ = store.Put(VisionCapability{ModelRef: visioncap.ModelRef(entry), Key: key, Status: visioncap.Probing, Attempts: current.Attempts})
	finishUnknown := func(reason error) (VisionCapability, error) {
		result := VisionCapability{
			ModelRef:  visioncap.ModelRef(entry),
			Key:       key,
			Status:    visioncap.Unknown,
			Reason:    "vision probe failed",
			Attempts:  current.Attempts + 1,
			CheckedAt: time.Now().UnixMilli(),
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
	ctx, cancel := context.WithTimeout(a.bootContext(), 25*time.Second)
	defer cancel()
	result := visioncap.Probe(ctx, prov, entry)
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
	return result, nil
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
		for i := range cfg.Providers {
			providerEntry := &cfg.Providers[i]
			if !providerEntry.Configured() {
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
