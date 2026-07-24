import { Bot, CalendarClock, FolderOpen, Library, Minus, PanelLeft, PanelRight, Search, Square, Sparkles, X } from "lucide-react";
import { useT } from "../lib/i18n";
import { Tooltip } from "./Tooltip";

type DesktopPlatform = "darwin" | "windows" | "linux";

interface AppChromeProps {
  platform: DesktopPlatform;
  browserPreviewChrome: boolean;
  sidebarTogglePressed: boolean;
  sidebarExpandBlocked: boolean;
  sidebarCollapsed: boolean;
  sidebarToggleTitle: string;
  workspacePanelMaximized: boolean;
  workspacePanelRenderable: boolean;
  workspaceTogglePressed: boolean;
  workspacePanelLabel: string;
  statusLights: Array<{ key: string; label: string; tone: "neutral" | "info" | "success" | "warn" | "danger"; active: boolean }>;
  onOpenProject: () => void;
  onOpenView: () => void;
  onOpenSkills: () => void;
  onOpenBots: () => void;
  onOpenAutomations: () => void;
  onOpenToolLibrary: () => void;
  onToggleSidebar: () => void;
  onToggleWorkspacePanel: () => void;
  onOpenPalette: () => void;
}

export function AppChrome({
  platform,
  browserPreviewChrome,
  sidebarTogglePressed,
  sidebarExpandBlocked,
  sidebarCollapsed,
  sidebarToggleTitle,
  workspacePanelMaximized,
  workspacePanelRenderable,
  workspaceTogglePressed,
  workspacePanelLabel,
  statusLights,
  onOpenProject,
  onOpenView,
  onOpenSkills,
  onOpenBots,
  onOpenAutomations,
  onOpenToolLibrary,
  onToggleSidebar,
  onToggleWorkspacePanel,
  onOpenPalette,
}: AppChromeProps) {
  const t = useT();
  const darwinChrome = platform === "darwin";
  const showWindowsPreviewControls = browserPreviewChrome && platform === "windows";
  const chromeClassName = [
    "app-chrome",
    "app-chrome--tabs",
    darwinChrome ? "app-chrome--darwin-tabs" : "app-chrome--native-tabs",
    !darwinChrome ? "app-chrome--identityless" : "",
    showWindowsPreviewControls ? "app-chrome--preview-window-controls" : "",
    `app-chrome--platform-${platform}`,
  ].filter(Boolean).join(" ");

  return (
    <header className={chromeClassName}>
      {browserPreviewChrome && darwinChrome && (
        <div className="app-chrome__traffic" aria-hidden="true">
          <span />
          <span />
          <span />
        </div>
      )}
      {darwinChrome && <span className="app-chrome__drag-rail" aria-hidden="true" />}
      <button
        className={[
          "app-chrome__panel-toggle",
          "app-chrome__panel-toggle--left",
          sidebarTogglePressed ? "app-chrome__panel-toggle--pressed" : "",
          sidebarExpandBlocked ? "app-chrome__panel-toggle--blocked" : "",
        ].filter(Boolean).join(" ")}
        type="button"
        onClick={sidebarExpandBlocked ? undefined : onToggleSidebar}
        aria-label={sidebarToggleTitle}
        aria-pressed={!sidebarCollapsed}
        aria-disabled={sidebarExpandBlocked}
      >
        <PanelLeft size={16} />
      </button>

      <div className="app-chrome__drag-surface" aria-hidden="true" />
      <nav className="app-chrome__actions" aria-label={t("topbar.actions")}>
        <Tooltip label={t("topbar.openProject")}>
          <button type="button" className="app-chrome__action app-chrome__action--core" onClick={onOpenProject} aria-label={t("topbar.openProject")}>
            <FolderOpen size={14} />
            <span>{t("topbar.openProject")}</span>
          </button>
        </Tooltip>
        <Tooltip label={t("topbar.view")}>
          <button type="button" className="app-chrome__action app-chrome__action--core" onClick={onOpenView} aria-label={t("topbar.view")}>
            <Square size={14} />
            <span>{t("topbar.view")}</span>
          </button>
        </Tooltip>
        <Tooltip label={t("topbar.skills")}>
          <button type="button" className="app-chrome__action app-chrome__action--secondary" onClick={onOpenSkills} aria-label={t("topbar.skills")}>
            <Sparkles size={14} />
            <span>{t("topbar.skills")}</span>
          </button>
        </Tooltip>
        <Tooltip label={t("topbar.botStatus")}>
          <button type="button" className="app-chrome__action app-chrome__action--secondary" onClick={onOpenBots} aria-label={t("topbar.botStatus")}>
            <Bot size={14} />
            <span>{t("topbar.bot")}</span>
          </button>
        </Tooltip>
        <Tooltip label={t("topbar.automationStatus")}>
          <button type="button" className="app-chrome__action app-chrome__action--tertiary" onClick={onOpenAutomations} aria-label={t("topbar.automationStatus")}>
            <CalendarClock size={14} />
            <span>{t("topbar.automation")}</span>
          </button>
        </Tooltip>
        <Tooltip label={t("topbar.toolLibraryStatus")}>
          <button type="button" className="app-chrome__action app-chrome__action--tertiary" onClick={onOpenToolLibrary} aria-label={t("topbar.toolLibraryStatus")}>
            <Library size={14} />
            <span>{t("topbar.toolLibrary")}</span>
          </button>
        </Tooltip>
      </nav>

      <div className="app-chrome__drag-surface app-chrome__drag-surface--center" aria-hidden="true" />

      <div className="app-chrome__status" aria-label={t("topbar.statusLights")}>
        {statusLights.map((light) => (
          <Tooltip key={light.key} label={light.label}>
            <span className={`app-chrome__status-light app-chrome__status-light--${light.tone}${light.active ? " app-chrome__status-light--active" : ""}`} />
          </Tooltip>
        ))}
      </div>

      <div className={[
          "app-chrome__tools",
          workspaceTogglePressed ? "app-chrome__tools--workspace-pressed" : "",
        ].filter(Boolean).join(" ")}
        aria-label={t("tabBar.commandSearch")}
      >
        <button
          className="tabbar__command tabbar__command--compact app-chrome__command"
          type="button"
          onClick={onOpenPalette}
          aria-label={t("palette.placeholder")}
        >
          <Search size={13} className="tabbar__command-icon" />
          <span className="tabbar__command-text tabbar__command-text--full">{t("tabBar.commandSearch")}</span>
          <span className="tabbar__command-text tabbar__command-text--compact">{t("tabBar.commandSearchCompact")}</span>
          <kbd className="tabbar__command-kbd">Ctrl+K</kbd>
        </button>
      </div>

      {!workspacePanelMaximized && (
        <button
          className={[
            "app-chrome__panel-toggle",
            "app-chrome__panel-toggle--right",
            workspacePanelRenderable ? "app-chrome__panel-toggle--active" : "",
            workspaceTogglePressed ? "app-chrome__panel-toggle--pressed" : "",
          ].filter(Boolean).join(" ")}
          type="button"
          onClick={onToggleWorkspacePanel}
          aria-label={workspacePanelLabel}
          aria-pressed={workspacePanelRenderable}
        >
          <PanelRight size={16} />
        </button>
      )}
      {showWindowsPreviewControls && (
        <div className="app-chrome__window-controls app-chrome__window-controls--windows" aria-hidden="true">
          <span className="app-chrome__window-control app-chrome__window-control--minimize">
            <Minus size={12} strokeWidth={1.9} />
          </span>
          <span className="app-chrome__window-control app-chrome__window-control--maximize">
            <Square size={10} strokeWidth={1.8} />
          </span>
          <span className="app-chrome__window-control app-chrome__window-control--close">
            <X size={12} strokeWidth={1.9} />
          </span>
        </div>
      )}
    </header>
  );
}
