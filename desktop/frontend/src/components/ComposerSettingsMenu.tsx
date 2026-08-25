import { useCallback, useEffect, useRef, useState } from "react";
import { Brain, Check, Gauge, Settings2, Shield, ShieldAlert, ShieldCheck, Sparkles } from "lucide-react";
import { asArray } from "../lib/array";
import { app } from "../lib/bridge";
import { useT } from "../lib/i18n";
import type { EffortInfo, ModelInfo, PromptMode, ToolApprovalMode } from "../lib/types";
import { ANCHORED_POPOVER_CLOSE_MS, AnchoredPopover } from "./AnchoredPopover";
import { Tooltip } from "./Tooltip";

function promptModeLabelKey(mode: PromptMode): "composer.promptMode.assistant" | "composer.promptMode.coding" {
  return mode === "assistant" ? "composer.promptMode.assistant" : "composer.promptMode.coding";
}

function approvalLabel(mode: ToolApprovalMode, t: ReturnType<typeof useT>): string {
  if (mode === "ask") return t("composer.modeAsk");
  if (mode === "auto") return t("composer.modeNormal");
  return t("composer.modeYolo");
}

function approvalIcon(mode: ToolApprovalMode) {
  if (mode === "ask") return <Shield size={15} />;
  if (mode === "auto") return <ShieldCheck size={15} />;
  return <ShieldAlert size={15} />;
}

export function ComposerSettingsMenu({
  tabId,
  disabled,
  modelLabel,
  effort,
  toolApprovalMode,
  showToolApprovalControls,
  promptMode,
  promptModes,
  promptModeLocked,
  promptModeSwitching,
  onSetToolApprovalMode,
  onSwitchModel,
  onSetEffort,
  onSetPromptMode,
}: {
  tabId?: string;
  disabled: boolean;
  modelLabel: string;
  effort?: EffortInfo;
  toolApprovalMode: ToolApprovalMode;
  showToolApprovalControls: boolean;
  promptMode: PromptMode;
  promptModes: PromptMode[];
  promptModeLocked: boolean;
  promptModeSwitching: boolean;
  onSetToolApprovalMode: (mode: ToolApprovalMode) => void;
  onSwitchModel: (modelRef: string, displayLabel?: string) => void;
  onSetEffort: (level: string) => void;
  onSetPromptMode: (mode: PromptMode) => void;
}) {
  const t = useT();
  const [open, setOpen] = useState(false);
  const [closing, setClosing] = useState(false);
  const [models, setModels] = useState<ModelInfo[]>([]);
  const [modelsLoading, setModelsLoading] = useState(false);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const closeTimerRef = useRef<number | null>(null);
  const levels = asArray(effort?.levels);
  const currentEffort = effort?.current || "auto";

  const clearCloseTimer = useCallback(() => {
    if (closeTimerRef.current === null) return;
    window.clearTimeout(closeTimerRef.current);
    closeTimerRef.current = null;
  }, []);

  const openMenu = useCallback(() => {
    clearCloseTimer();
    setClosing(false);
    setOpen(true);
  }, [clearCloseTimer]);

  const closeMenu = useCallback((afterClose?: () => void) => {
    clearCloseTimer();
    setClosing(true);
    window.requestAnimationFrame(() => setOpen(false));
    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    closeTimerRef.current = window.setTimeout(() => {
      closeTimerRef.current = null;
      setClosing(false);
      afterClose?.();
    }, reduceMotion ? 0 : ANCHORED_POPOVER_CLOSE_MS);
  }, [clearCloseTimer]);

  useEffect(() => () => clearCloseTimer(), [clearCloseTimer]);

  useEffect(() => {
    if (!open) return;
    let cancelled = false;
    setModelsLoading(true);
    const request = tabId ? app.ModelsForTab(tabId) : app.Models();
    request
      .then((next) => {
        if (!cancelled) setModels(asArray(next));
      })
      .catch(() => {
        if (!cancelled) setModels([]);
      })
      .finally(() => {
        if (!cancelled) setModelsLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [open, tabId]);

  const pick = (action: () => void) => {
    closeMenu(action);
  };

  const summary = [modelLabel, effort?.supported ? currentEffort : "", t(promptModeLabelKey(promptMode))].filter(Boolean).join(" · ");

  return (
    <div className="composer-session-settings">
      <Tooltip label={t("composer.sessionSettings")} disabled={open || closing}>
        <button
          ref={triggerRef}
          type="button"
          className={"composer-session-settings__trigger" + (open || closing ? " composer-session-settings__trigger--open" : "")}
          onClick={() => (open || closing ? closeMenu() : openMenu())}
          disabled={disabled}
          aria-haspopup="menu"
          aria-expanded={open && !closing}
          aria-busy={promptModeSwitching}
          aria-label={t("composer.sessionSettings")}
          title={summary}
        >
          <Settings2 size={16} />
        </button>
      </Tooltip>
      <AnchoredPopover
        open={open && !disabled}
        closing={closing}
        anchorRef={triggerRef}
        onClose={() => closeMenu()}
        className="composer-access-menu composer-session-settings__menu"
        align="start"
      >
        {showToolApprovalControls && (
          <div className="composer-access-menu__section" role="radiogroup" aria-label={t("composer.accessMenuTitle")}>
            <div className="composer-access-menu__label">{t("composer.accessMenuTitle")}</div>
            <div className="composer-session-settings__grid composer-session-settings__grid--three">
              {(["ask", "auto", "yolo"] as ToolApprovalMode[]).map((mode) => (
                <button
                  key={mode}
                  type="button"
                  role="radio"
                  aria-checked={toolApprovalMode === mode}
                  className={"composer-session-settings__choice" + (toolApprovalMode === mode ? " composer-session-settings__choice--active" : "")}
                  onClick={() => pick(() => onSetToolApprovalMode(mode))}
                >
                  {approvalIcon(mode)}
                  <span>{approvalLabel(mode, t)}</span>
                </button>
              ))}
            </div>
          </div>
        )}
        <div className="composer-access-menu__section">
          <div className="composer-access-menu__label">{t("status.modelTitle")}</div>
          <div className="composer-session-settings__models" role="listbox" aria-label={t("status.modelTitle")}>
            {modelsLoading && models.length === 0 && <div className="composer-session-settings__empty">{t("common.loading")}</div>}
            {!modelsLoading && models.length === 0 && <div className="composer-session-settings__empty">{t("status.noModels")}</div>}
            {models.map((model) => (
              <button
                key={model.ref}
                type="button"
                role="option"
                aria-selected={model.current}
                className={"composer-session-settings__model" + (model.current ? " composer-session-settings__model--active" : "")}
                onClick={() => pick(() => {
                  if (!model.current) onSwitchModel(model.ref, model.model);
                })}
              >
                <Brain size={14} />
                <span className="composer-session-settings__model-copy">
                  <span title={model.model}>{model.model}</span>
                  <small title={model.provider}>{model.provider}</small>
                </span>
                {model.current && <Check size={13} />}
              </button>
            ))}
          </div>
        </div>
        {effort?.supported && levels.length > 0 && (
          <div className="composer-access-menu__section">
            <div className="composer-access-menu__label">{t("status.effortTitle")}</div>
            <div className="composer-session-settings__grid" role="listbox" aria-label={t("status.effortTitle")}>
              {levels.map((level) => (
                <button
                  key={level}
                  type="button"
                  role="option"
                  aria-selected={level === currentEffort}
                  className={"composer-session-settings__choice" + (level === currentEffort ? " composer-session-settings__choice--active" : "")}
                  onClick={() => pick(() => {
                    if (level !== currentEffort) onSetEffort(level);
                  })}
                >
                  <Gauge size={14} />
                  <span>{level}</span>
                </button>
              ))}
            </div>
          </div>
        )}
        {promptModes.length > 0 && (
          <div className="composer-access-menu__section">
            <div className="composer-access-menu__label">{t("composer.promptMode.title")}</div>
            <div className="composer-session-settings__grid" role="listbox" aria-label={t("composer.promptMode.title")}>
              {promptModes.map((mode) => (
                <button
                  key={mode}
                  type="button"
                  role="option"
                  aria-selected={mode === promptMode}
                  className={"composer-session-settings__choice composer-session-settings__choice--" + mode + (mode === promptMode ? " composer-session-settings__choice--active" : "")}
                  onClick={() => pick(() => {
                    if (mode !== promptMode) onSetPromptMode(mode);
                  })}
                  disabled={promptModeLocked}
                >
                  <Sparkles size={14} />
                  <span>{t(promptModeLabelKey(mode))}</span>
                </button>
              ))}
            </div>
          </div>
        )}
      </AnchoredPopover>
    </div>
  );
}
