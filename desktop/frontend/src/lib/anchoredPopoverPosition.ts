export type PopoverAlign = "start" | "center" | "end";
export type PopoverPlacement = "auto" | "bottom";

type RectLike = Pick<DOMRect, "left" | "right" | "top" | "bottom" | "width" | "height">;

export type AnchoredPopoverPosition = {
  left: number;
  top: number;
};

function clamp(value: number, min: number, max: number): number {
  return Math.min(Math.max(value, min), max);
}

export function calculateAnchoredPopoverPosition({
  anchor,
  menu,
  align,
  offset,
  placement,
  viewportWidth,
  viewportHeight,
  edgeGap = 8,
}: {
  anchor: RectLike;
  menu: RectLike;
  align: PopoverAlign;
  offset: number;
  placement: PopoverPlacement;
  viewportWidth: number;
  viewportHeight: number;
  edgeGap?: number;
}): AnchoredPopoverPosition {
  const preferredTop = anchor.top - menu.height - offset;
  const fallbackTop = anchor.bottom + offset;
  const top = placement === "bottom"
    ? Math.min(fallbackTop, Math.max(edgeGap, viewportHeight - menu.height - edgeGap))
    : preferredTop >= edgeGap
    ? preferredTop
    : Math.min(fallbackTop, Math.max(edgeGap, viewportHeight - menu.height - edgeGap));
  const rawLeft = align === "end"
    ? anchor.right - menu.width
    : align === "center"
    ? anchor.left + (anchor.width - menu.width) / 2
    : anchor.left;
  const left = clamp(rawLeft, edgeGap, Math.max(edgeGap, viewportWidth - menu.width - edgeGap));
  return {
    left,
    top: clamp(top, edgeGap, Math.max(edgeGap, viewportHeight - menu.height - edgeGap)),
  };
}
