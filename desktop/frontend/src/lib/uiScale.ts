export const UI_SCALE_AUTO = 0;
const UI_SCALE_KEY = "orca-ui-scale";

export function isUIScale(value: unknown): value is number {
  return value === UI_SCALE_AUTO;
}

export function effectiveUIScale(_preference: number): number {
  return 100;
}

export function getUIScale(): number {
  return UI_SCALE_AUTO;
}

export function applyUIScale(_preference: number): number {
  if (typeof document !== "undefined") {
    const root = document.documentElement;
    root.style.removeProperty("zoom");
    delete root.dataset.uiScale;
    delete root.dataset.uiScaleEffective;
  }
  try {
    localStorage.removeItem(UI_SCALE_KEY);
  } catch {
    // Storage may be unavailable in isolated webviews.
  }
  return 100;
}

export function initUIScale(): void {
  applyUIScale(getUIScale());
}
