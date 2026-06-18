import { useCallback, useEffect, useRef, useState } from "react";
import { Eraser, Send, Square } from "lucide-react";
import { app } from "../lib/bridge";
import { useT } from "../lib/i18n";
import type { SideChatMessage } from "../lib/types";

export function SideChatPanel({ tabId }: { tabId?: string }) {
  const t = useT();
  const [messages, setMessages] = useState<SideChatMessage[]>([]);
  const [input, setInput] = useState("");
  const [busy, setBusy] = useState(false);
  const [err, setErr] = useState("");
  const listRef = useRef<HTMLDivElement>(null);

  const load = useCallback(async () => {
    if (!tabId) {
      setMessages([]);
      return;
    }
    setErr("");
    try {
      setMessages(await app.ListSideChat(tabId));
    } catch (e) {
      setErr(String((e as Error)?.message ?? e));
    }
  }, [tabId]);

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    const el = listRef.current;
    if (el) el.scrollTop = el.scrollHeight;
  }, [messages.length, busy]);

  const send = async () => {
    const text = input.trim();
    if (!tabId || !text || busy) return;
    setInput("");
    setBusy(true);
    setErr("");
    const optimistic: SideChatMessage = { id: `pending-${Date.now()}`, role: "user", content: text, createdAt: Date.now() };
    setMessages((current) => [...current, optimistic]);
    try {
      await app.SendSideChat(tabId, text);
      await load();
    } catch (e) {
      setErr(String((e as Error)?.message ?? e));
      await load();
    } finally {
      setBusy(false);
    }
  };

  const stop = async () => {
    if (!tabId) return;
    await app.CancelSideChat(tabId).catch(() => {});
    setBusy(false);
    await load();
  };

  const clear = async () => {
    if (!tabId || busy) return;
    await app.ClearSideChat(tabId);
    setMessages([]);
  };

  return (
    <section className="side-chat" aria-label={t("sideChat.title")}>
      <header className="side-chat__head">
        <div>
          <h2>{t("sideChat.title")}</h2>
          <p>{t("sideChat.summary")}</p>
        </div>
        <button type="button" className="side-chat__icon" disabled={!tabId || busy || messages.length === 0} onClick={() => void clear()} aria-label={t("sideChat.clear")}>
          <Eraser size={14} />
        </button>
      </header>
      {err && <div className="side-chat__error">{err}</div>}
      <div className="side-chat__messages" ref={listRef}>
        {messages.length === 0 ? (
          <div className="side-chat__empty">{t("sideChat.empty")}</div>
        ) : (
          messages.map((message) => (
            <div className={`side-chat__msg side-chat__msg--${message.role}`} key={message.id}>
              <div className="side-chat__bubble">{message.content}</div>
            </div>
          ))
        )}
        {busy && (
          <div className="side-chat__msg side-chat__msg--assistant">
            <div className="side-chat__bubble side-chat__bubble--loading">{t("sideChat.loading")}</div>
          </div>
        )}
      </div>
      <form
        className="side-chat__composer"
        onSubmit={(event) => {
          event.preventDefault();
          void send();
        }}
      >
        <textarea
          value={input}
          disabled={!tabId}
          placeholder={t("sideChat.placeholder")}
          onChange={(event) => setInput(event.target.value)}
          onKeyDown={(event) => {
            if (event.key === "Enter" && !event.shiftKey) {
              event.preventDefault();
              void send();
            }
          }}
        />
        {busy ? (
          <button type="button" className="side-chat__send" onClick={() => void stop()} aria-label={t("sideChat.stop")}>
            <Square size={14} />
          </button>
        ) : (
          <button type="submit" className="side-chat__send" disabled={!input.trim() || !tabId} aria-label={t("sideChat.send")}>
            <Send size={14} />
          </button>
        )}
      </form>
    </section>
  );
}
