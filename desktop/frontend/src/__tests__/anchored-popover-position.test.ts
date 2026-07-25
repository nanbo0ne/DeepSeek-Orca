import { calculateAnchoredPopoverPosition } from "../lib/anchoredPopoverPosition";

function eq(actual: number, expected: number, message: string) {
  if (actual !== expected) throw new Error(`${message}: got ${actual}, want ${expected}`);
}

const anchor = { left: 400, right: 600, top: 500, bottom: 534, width: 200, height: 34 };
const menu = { left: 0, right: 400, top: 0, bottom: 180, width: 400, height: 180 };

const centered = calculateAnchoredPopoverPosition({
  anchor,
  menu,
  align: "center",
  offset: 6,
  placement: "auto",
  viewportWidth: 1200,
  viewportHeight: 800,
});
eq(centered.left, 300, "centers the popover over its anchor");
eq(centered.top, 314, "places the popover above its anchor");

const edgeClamped = calculateAnchoredPopoverPosition({
  anchor: { ...anchor, left: 4, right: 204 },
  menu,
  align: "center",
  offset: 6,
  placement: "auto",
  viewportWidth: 1200,
  viewportHeight: 800,
});
eq(edgeClamped.left, 8, "keeps a centered popover inside the viewport");

console.log("2 passed, 0 failed, 2 total");
