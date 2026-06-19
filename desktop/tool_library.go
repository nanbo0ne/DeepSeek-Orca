package main

import (
	"fmt"

	"deepseek-orca/internal/config"
)

type ToolLibrarySettings = config.ToolLibraryConfig

func (a *App) GetToolLibrarySettings() (ToolLibrarySettings, error) {
	cfg, err := config.LoadForRoot(a.activeWorkspaceRoot())
	if err != nil {
		return config.DefaultToolLibrarySettings(), err
	}
	return cfg.ToolLibrary, nil
}

func (a *App) SetToolLibrarySettings(settings ToolLibrarySettings) error {
	cfg, path, err := a.loadDesktopUserConfigForEdit()
	if err != nil {
		return err
	}
	cfg.ToolLibrary = config.NormalizeToolLibrarySettings(settings)
	if err := cfg.SaveTo(path); err != nil {
		return err
	}
	tab := a.activeTab()
	if tab == nil {
		return nil
	}
	if tab.Ctrl != nil && tab.Ctrl.Running() {
		a.noticeForTab(tab.ID, "工具库设置已保存，将从下一轮对话生效。")
		return nil
	}
	if err := a.rebuild(); err != nil {
		return fmt.Errorf("tool library saved, but rebuilding the current session failed: %w", err)
	}
	return nil
}
