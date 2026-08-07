import { useCallback, useEffect, useReducer, useRef } from "react";
import { Check, ChevronUp, Circle, CircleDot, Pin, X } from "lucide-react";
import { useT } from "../lib/i18n";
import { createTodoPanelState, isTodoPanelOpen, reduceTodoPanelState } from "../lib/todoPanelState";
import type { Todo } from "../lib/tools";
import { Tooltip } from "./Tooltip";

const CLOSE_DELAY_MS = 180;

export function TodoPanel({ todoId, todos, onDismiss }: { todoId: string; todos: Todo[]; onDismiss: () => void }) {
  const t = useT();
  const [panel, dispatch] = useReducer(reduceTodoPanelState, todoId, createTodoPanelState);
  const currentRef = useRef<HTMLLIElement | null>(null);
  const closeTimerRef = useRef<number | null>(null);
  const open = isTodoPanelOpen(panel);

  const done = todos.filter((todo) => todo.status === "completed").length;
  const current = todos.find((todo) => todo.status === "in_progress");
  const runningIndex = todos.findIndex((todo) => todo.status === "in_progress");
  const pendingIndex = todos.findIndex((todo) => todo.status !== "completed");
  const activeIndex = runningIndex >= 0 ? runningIndex + 1 : pendingIndex >= 0 ? pendingIndex + 1 : todos.length;
  const activeTodo = current ?? (pendingIndex >= 0 ? todos[pendingIndex] : todos[todos.length - 1]);
  const activeText = activeTodo?.status === "in_progress" && activeTodo.activeForm
    ? activeTodo.activeForm
    : activeTodo?.content ?? "";
  const progressText = t("todo.progress", { current: activeIndex, total: todos.length });

  const cancelClose = useCallback(() => {
    if (closeTimerRef.current !== null) window.clearTimeout(closeTimerRef.current);
    closeTimerRef.current = null;
  }, []);

  const scheduleTransientClose = useCallback(() => {
    cancelClose();
    closeTimerRef.current = window.setTimeout(() => {
      dispatch({ type: "hover", value: false });
      dispatch({ type: "focus", value: false });
    }, CLOSE_DELAY_MS);
  }, [cancelClose]);

  useEffect(() => {
    dispatch({ type: "list", listId: todoId });
  }, [todoId]);

  useEffect(() => () => cancelClose(), [cancelClose]);

  useEffect(() => {
    if (open) currentRef.current?.scrollIntoView({ block: "nearest" });
  }, [open, current?.content, current?.activeForm]);

  if (todos.length === 0) return null;

  return (
    <div
      className={`todobar${open ? " todobar--open" : ""}`}
      onMouseEnter={() => { cancelClose(); dispatch({ type: "hover", value: true }); }}
      onMouseLeave={scheduleTransientClose}
      onFocus={() => { cancelClose(); dispatch({ type: "focus", value: true }); }}
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget as Node | null)) scheduleTransientClose();
      }}
      onKeyDown={(event) => {
        if (event.key === "Escape" && open) {
          event.preventDefault();
          dispatch({ type: "close" });
        }
      }}
    >
      <section
        className="todobar__surface"
        data-ui-surface="panel"
        aria-label={t("todo.title")}
      >
        <button
          type="button"
          className={`todobar__trigger${panel.pinned ? " todobar__trigger--pinned" : ""}`}
          aria-expanded={open}
          aria-controls="todo-details"
          aria-label={`${activeText}. ${progressText}`}
          title={activeText}
          onClick={() => dispatch({ type: "toggle-pin" })}
        >
          <span className="todobar__progress-track" aria-hidden="true">
            <span className="todobar__progress-fill" style={{ width: `${Math.round((done / todos.length) * 100)}%` }} />
          </span>
          <span className="todobar__active-text">{activeText}</span>
          <span className="todobar__progress-label" aria-hidden="true">{activeIndex}/{todos.length}</span>
          <ChevronUp size={13} aria-hidden="true" />
        </button>

        {open && (
          <div id="todo-details" className="todobar__details">
            <header className="todobar__head">
              <div className="todobar__heading">
                <span className="todobar__title">{t("todo.title")}</span>
                <span className="todobar__count">{done}/{todos.length}</span>
              </div>
              <div className="todobar__actions">
                {panel.pinned && <Pin size={12} className="todobar__pin" aria-label={t("todo.pinned")} />}
                <Tooltip label={t("todo.dismiss")}>
                  <button type="button" className="todobar__close" onClick={onDismiss}>
                    <X size={13} />
                  </button>
                </Tooltip>
              </div>
            </header>
            <ul className="todobar__list">
              {todos.map((todo, index) => (
                <li
                  key={`${todo.content}-${index}`}
                  ref={todo.status === "in_progress" ? currentRef : undefined}
                  className={`todobar__item todobar__item--${todo.status}${todo.level ? " todobar__item--sub" : ""}`}
                >
                  {todo.status === "completed" ? (
                    <Check size={14} className="todobar__ico todobar__ico--done" />
                  ) : todo.status === "in_progress" ? (
                    <CircleDot size={14} className="todobar__ico todobar__ico--active" />
                  ) : (
                    <Circle size={14} className="todobar__ico" />
                  )}
                  <span className="todobar__text">
                    {todo.status === "in_progress" && todo.activeForm ? todo.activeForm : todo.content}
                  </span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </section>
    </div>
  );
}
