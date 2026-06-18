import { useCallback, useEffect, useState } from "react";
import { Pause, Play, RefreshCw, XCircle } from "lucide-react";
import { app } from "../lib/bridge";
import { useT } from "../lib/i18n";
import type { AutomationView } from "../lib/types";
import { ModalCloseButton } from "./ModalCloseButton";

export function AutomationPanel({ onClose }: { onClose: () => void }) {
  const t = useT();
  const [items, setItems] = useState<AutomationView[]>([]);
  const [busy, setBusy] = useState("");
  const [err, setErr] = useState("");

  const load = useCallback(async () => {
    setErr("");
    try {
      setItems(await app.ListAutomations());
    } catch (e) {
      setErr(String((e as Error)?.message ?? e));
    }
  }, []);

  useEffect(() => {
    void load();
  }, [load]);

  const mutate = async (id: string, fn: () => Promise<void>) => {
    setBusy(id);
    setErr("");
    try {
      await fn();
      await load();
    } catch (e) {
      setErr(String((e as Error)?.message ?? e));
    } finally {
      setBusy("");
    }
  };

  return (
    <div className="management-modal-backdrop automation-modal-backdrop" onClick={(event) => { if (event.target === event.currentTarget) onClose(); }}>
      <div className="management-modal automation-modal">
        <header className="management-modal__head">
          <div>
            <div className="management-modal__title">{t("automation.title")}</div>
            <div className="management-modal__summary">{t("automation.summary")}</div>
          </div>
          <div className="management-modal__actions">
            <button type="button" className="automation-panel__icon" onClick={() => void load()} aria-label={t("automation.refresh")}>
              <RefreshCw size={14} />
            </button>
            <button type="button" className="automation-panel__clear" onClick={() => void mutate("__clear__", () => app.ClearFinishedAutomations())}>
              {t("automation.clearFinished")}
            </button>
            <ModalCloseButton label={t("common.close")} onClick={onClose} />
          </div>
        </header>
        {err && <div className="banner banner--error">{err}</div>}
        <div className="automation-panel">
          {items.length === 0 ? (
            <div className="automation-panel__empty">{t("automation.empty")}</div>
          ) : (
            items.map((item) => (
              <article className="automation-panel__item" key={item.id}>
                <div className="automation-panel__main">
                  <div className="automation-panel__title-row">
                    <strong>{item.label || item.id}</strong>
                    <span className={`automation-panel__status automation-panel__status--${item.status}`}>{automationStatusLabel(item.status, t)}</span>
                  </div>
                  <div className="automation-panel__meta">{item.schedule || item.kind} · {automationActionLabel(item.action, t)}</div>
                  {item.nextRunAt && <div className="automation-panel__sub">{t("automation.nextRun")}: {formatTime(item.nextRunAt)}</div>}
                  {item.lastRunAt && <div className="automation-panel__sub">{t("automation.lastRun")}: {formatTime(item.lastRunAt)}</div>}
                  {(item.error || item.result) && <pre className="automation-panel__result">{item.error || item.result}</pre>}
                </div>
                <div className="automation-panel__actions">
                  {item.status === "paused" ? (
                    <button type="button" disabled={busy === item.id} onClick={() => void mutate(item.id, () => app.ResumeAutomation(item.id))}>
                      <Play size={13} />
                      <span>{t("automation.resume")}</span>
                    </button>
                  ) : (
                    <button type="button" disabled={busy === item.id || item.status !== "scheduled"} onClick={() => void mutate(item.id, () => app.PauseAutomation(item.id))}>
                      <Pause size={13} />
                      <span>{t("automation.pause")}</span>
                    </button>
                  )}
                  <button type="button" disabled={busy === item.id || item.status === "cancelled"} onClick={() => void mutate(item.id, () => app.CancelAutomation(item.id))}>
                    <XCircle size={13} />
                    <span>{t("automation.cancel")}</span>
                  </button>
                </div>
              </article>
            ))
          )}
        </div>
      </div>
    </div>
  );
}

function automationStatusLabel(status: string, t: ReturnType<typeof useT>): string {
  switch (status) {
    case "running": return t("automation.status.running");
    case "scheduled": return t("automation.status.scheduled");
    case "paused": return t("automation.status.paused");
    case "failed": return t("automation.status.failed");
    case "cancelled": return t("automation.status.cancelled");
    default: return status || t("common.none");
  }
}

function automationActionLabel(action: string, t: ReturnType<typeof useT>): string {
  if (action === "notify") return t("automation.action.notify");
  if (action === "host_command") return t("automation.action.hostCommand");
  return action || t("common.none");
}

function formatTime(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return date.toLocaleString();
}
