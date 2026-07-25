import { useEffect, useLayoutEffect, useRef, useState } from "react";
import type { CSSProperties, ReactNode, RefObject } from "react";
import { createPortal } from "react-dom";
import {
  calculateAnchoredPopoverPosition,
  type PopoverAlign,
  type PopoverPlacement,
} from "../lib/anchoredPopoverPosition";

type PopoverPosition = {
  left: number;
  top: number;
};
type PopoverPhase = "closed" | "open" | "closing";

const EDGE_GAP = 8;
const DEFAULT_OFFSET = 8;
export const ANCHORED_POPOVER_CLOSE_MS = 140;

function samePosition(a: PopoverPosition | null, b: PopoverPosition): boolean {
  return !!a && Math.abs(a.left - b.left) < 0.5 && Math.abs(a.top - b.top) < 0.5;
}

export function AnchoredPopover({
  open,
  anchorRef,
  onClose,
  className,
  children,
  align = "start",
  offset = DEFAULT_OFFSET,
  placement = "auto",
  style,
  closing = false,
}: {
  open: boolean;
  anchorRef: RefObject<HTMLElement | null>;
  onClose: () => void;
  className: string;
  children: ReactNode;
  align?: PopoverAlign;
  offset?: number;
  placement?: PopoverPlacement;
  style?: CSSProperties;
  closing?: boolean;
}) {
  const [phase, setPhase] = useState<PopoverPhase>(open ? "open" : "closed");
  const [position, setPosition] = useState<PopoverPosition | null>(null);
  const popoverRef = useRef<HTMLDivElement>(null);
  const phaseRef = useRef<PopoverPhase>(phase);

  useLayoutEffect(() => {
    let id: number | undefined;
    if (open) {
      phaseRef.current = "open";
      setPhase("open");
      return undefined;
    }
    if (phaseRef.current === "closed") return undefined;
    phaseRef.current = "closing";
    setPhase("closing");
    const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    id = window.setTimeout(() => {
      phaseRef.current = "closed";
      setPhase("closed");
      setPosition(null);
    }, reduceMotion ? 0 : ANCHORED_POPOVER_CLOSE_MS);
    return () => {
      if (id !== undefined) window.clearTimeout(id);
    };
  }, [open]);

  const rendered = closing || phase !== "closed";

  useLayoutEffect(() => {
    if (!rendered) {
      setPosition(null);
      return;
    }
    let frame: number | null = null;
    const updatePosition = () => {
      frame = null;
      const anchor = anchorRef.current?.getBoundingClientRect();
      const menu = popoverRef.current?.getBoundingClientRect();
      if (!anchor || !menu) return;
      const next = calculateAnchoredPopoverPosition({
        anchor,
        menu,
        align,
        offset,
        placement,
        viewportWidth: window.innerWidth,
        viewportHeight: window.innerHeight,
        edgeGap: EDGE_GAP,
      });
      setPosition((current) => (samePosition(current, next) ? current : next));
    };
    const scheduleUpdate = () => {
      if (frame !== null) return;
      frame = window.requestAnimationFrame(updatePosition);
    };
    updatePosition();
    scheduleUpdate();

    const anchor = anchorRef.current;
    const menu = popoverRef.current;
    let observer: ResizeObserver | null = null;
    if (typeof ResizeObserver !== "undefined") {
      observer = new ResizeObserver(scheduleUpdate);
      if (anchor) observer.observe(anchor);
      if (menu) observer.observe(menu);
    }

    return () => {
      if (frame !== null) window.cancelAnimationFrame(frame);
      observer?.disconnect();
    };
  }, [rendered, anchorRef, align, offset, placement]);

  useEffect(() => {
    if (!open || closing) return;
    const closeOnEscape = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    const closeOnOutsideClick = (event: MouseEvent) => {
      const target = event.target;
      if (!(target instanceof Node)) return;
      if (popoverRef.current?.contains(target) || anchorRef.current?.contains(target)) return;
      onClose();
    };
    const closeOnViewportChange = () => onClose();
    window.addEventListener("keydown", closeOnEscape);
    document.addEventListener("click", closeOnOutsideClick);
    window.addEventListener("resize", closeOnViewportChange);
    return () => {
      window.removeEventListener("keydown", closeOnEscape);
      document.removeEventListener("click", closeOnOutsideClick);
      window.removeEventListener("resize", closeOnViewportChange);
    };
  }, [anchorRef, onClose, open]);

  if (!rendered) return null;

  return createPortal(
    <div
      ref={popoverRef}
      data-anchored-popover="active"
      data-ready={position ? "true" : "false"}
      data-state={closing || phase === "closing" ? "closing" : "open"}
      aria-hidden={closing || phase === "closing" ? true : undefined}
      className={`anchored-popover ${className}`}
      style={{
        ...style,
        left: position?.left ?? -9999,
        top: position?.top ?? -9999,
        visibility: position ? "visible" : "hidden",
      }}
      onMouseDown={(event) => {
        event.stopPropagation();
      }}
      onClick={(event) => {
        event.stopPropagation();
      }}
    >
      {children}
    </div>,
    document.body,
  );
}
