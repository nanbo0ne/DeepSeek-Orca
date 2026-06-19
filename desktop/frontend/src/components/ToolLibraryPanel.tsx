import { useCallback, useEffect, useMemo, useState } from "react";
import { RefreshCw } from "lucide-react";
import { app } from "../lib/bridge";
import { useT } from "../lib/i18n";
import type { ToolLibrarySettings } from "../lib/types";
import type { DictKey } from "../locales/en";
import { ModalCloseButton } from "./ModalCloseButton";

const defaultSettings: ToolLibrarySettings = {
  threadManagementEnabled: true,
  webSearchEnabled: true,
  replRuntimeEnabled: true,
  documentToolsEnabled: true,
  hostSystemToolsEnabled: true,
  conversationSearchEnabled: true,
  proactiveToolUseEnabled: true,
};

type ToolLibraryKey = keyof ToolLibrarySettings;

const rows: Array<{ key: ToolLibraryKey; title: DictKey; desc: DictKey }> = [
  { key: "threadManagementEnabled", title: "toolLibrary.threadManagement", desc: "toolLibrary.threadManagementDesc" },
  { key: "webSearchEnabled", title: "toolLibrary.webSearch", desc: "toolLibrary.webSearchDesc" },
  { key: "replRuntimeEnabled", title: "toolLibrary.replRuntime", desc: "toolLibrary.replRuntimeDesc" },
  { key: "documentToolsEnabled", title: "toolLibrary.documentTools", desc: "toolLibrary.documentToolsDesc" },
  { key: "hostSystemToolsEnabled", title: "toolLibrary.hostSystemTools", desc: "toolLibrary.hostSystemToolsDesc" },
  { key: "conversationSearchEnabled", title: "toolLibrary.conversationSearch", desc: "toolLibrary.conversationSearchDesc" },
  { key: "proactiveToolUseEnabled", title: "toolLibrary.proactiveToolUse", desc: "toolLibrary.proactiveToolUseDesc" },
];

export function ToolLibraryPanel({ onClose }: { onClose: () => void }) {
  const t = useT();
  const [settings, setSettings] = useState<ToolLibrarySettings>(defaultSettings);
  const [saved, setSaved] = useState<ToolLibrarySettings>(defaultSettings);
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");
  const [savedHint, setSavedHint] = useState("");

  const load = useCallback(async () => {
    setBusy(true);
    setErr("");
    setSavedHint("");
    try {
      const next = normalizeSettings(await app.GetToolLibrarySettings());
      setSettings(next);
      setSaved(next);
    } catch (e) {
      setErr(String((e as Error)?.message ?? e));
    } finally {
      setBusy(false);
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const dirty = useMemo(() => JSON.stringify(settings) !== JSON.stringify(saved), [settings, saved]);

  const save = async () => {
    setBusy(true);
    setErr("");
    setSavedHint("");
    try {
      await app.SetToolLibrarySettings(settings);
      setSaved(settings);
      setSavedHint(t("toolLibrary.saved"));
    } catch (e) {
      setErr(String((e as Error)?.message ?? e));
    } finally {
      setBusy(false);
    }
  };

  const toggle = (key: ToolLibraryKey) => {
    setSettings((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  return (
    <div className="management-modal-backdrop tool-library-modal-backdrop" onClick={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <div className="management-modal tool-library-modal">
        <header className="management-modal__head">
          <div>
            <div className="management-modal__title">{t("toolLibrary.title")}</div>
            <div className="management-modal__summary">{t("toolLibrary.summary")}</div>
          </div>
          <div className="management-modal__actions">
            <button type="button" className="automation-panel__icon" disabled={busy} onClick={() => void load()} aria-label={t("toolLibrary.refresh")}>
              <RefreshCw size={14} />
            </button>
            <ModalCloseButton label={t("common.close")} onClick={onClose} />
          </div>
        </header>
        {err && <div className="banner banner--error">{err}</div>}
        {savedHint && <div className="banner banner--success">{savedHint}</div>}
        <div className="tool-library-panel">
          {rows.map((row) => (
            <button
              type="button"
              key={row.key}
              className={`tool-library-panel__row${settings[row.key] ? " tool-library-panel__row--on" : ""}`}
              onClick={() => toggle(row.key)}
              aria-pressed={settings[row.key]}
            >
              <span className="tool-library-panel__copy">
                <strong>{t(row.title)}</strong>
                <span>{t(row.desc)}</span>
              </span>
              <span className="tool-library-panel__switch" aria-hidden="true">
                <span />
              </span>
            </button>
          ))}
        </div>
        <footer className="tool-library-panel__footer">
          <span>{t("toolLibrary.effectHint")}</span>
          <div className="tool-library-panel__actions">
            <button type="button" className="btn btn--secondary btn--small" disabled={busy || !dirty} onClick={() => { setSettings(saved); setSavedHint(""); }}>
              {t("common.cancel")}
            </button>
            <button type="button" className="btn btn--primary btn--small" disabled={busy || !dirty} onClick={() => void save()}>
              {t("common.save")}
            </button>
          </div>
        </footer>
      </div>
    </div>
  );
}

function normalizeSettings(value: ToolLibrarySettings | null | undefined): ToolLibrarySettings {
  return { ...defaultSettings, ...(value ?? {}) };
}
