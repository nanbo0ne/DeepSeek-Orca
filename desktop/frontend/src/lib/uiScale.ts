export const UI_SCALE_AUTO = 0;
export const UI_SCALE_MIN = 80;
export const UI_SCALE_MAX = 125;
export const UI_SCALE_STEP = 5;

const UI_SCALE_KEY = "deepseek-orca-ui-scale";

export function isUIScale(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && (value === UI_SCALE_AUTO || (value >= UI_SCALE_MIN && value <= UI_SCALE_MAX && value % UI_SCALE_STEP === 0));
}

export function automaticUIScale(width: number, height: number): number {
  if (!Number.isFinite(width) || !Number.isFinite(height) || width <= 0 || height <= 0) return 100;
  const ratio = Math.min(width / 1920, height / 1080);
  const rounded = Math.round((ratio * 100) / UI_SCALE_STEP) * UI_SCALE_STEP;
  return Math.min(110, Math.max(90, rounded));
}

export function currentScreenSize(): { width: number; height: number } {
  if (typeof window === "undefined") return { width: 1920, height: 1080 };
  return {
    width: window.screen?.availWidth || window.innerWidth || 1920,
    height: window.screen?.availHeight || window.innerHeight || 1080,
  };
}

export function effectiveUIScale(preference: number, size = currentScreenSize()): number {
  return preference === UI_SCALE_AUTO ? automaticUIScale(size.width, size.height) : isUIScale(preference) ? preference : 100;
}

export function getUIScale(): number {
  if (typeof localStorage === "undefined") return UI_SCALE_AUTO;
  const parsed = Number.parseInt(localStorage.getItem(UI_SCALE_KEY) ?? "", 10);
  return isUIScale(parsed) ? parsed : UI_SCALE_AUTO;
}

export function applyUIScale(preference: number): number {
  const normalized = isUIScale(preference) ? preference : UI_SCALE_AUTO;
  const effective = effectiveUIScale(normalized);
  if (typeof document !== "undefined") {
    const root = document.documentElement;
    root.style.zoom = String(effective / 100);
    root.dataset.uiScale = normalized === UI_SCALE_AUTO ? "auto" : String(normalized);
    root.dataset.uiScaleEffective = String(effective);
    root.dispatchEvent(new CustomEvent("orca:ui-scale", { detail: { preference: normalized, effective } }));
  }
  try {
    localStorage.setItem(UI_SCALE_KEY, String(normalized));
  } catch {
    // The DOM value remains effective when persistent storage is unavailable.
  }
  return effective;
}

export function initUIScale(): void {
  applyUIScale(getUIScale());
  if (typeof window === "undefined") return;
  let lastWidth = window.screen?.availWidth;
  let lastHeight = window.screen?.availHeight;
  window.addEventListener("resize", () => {
    const width = window.screen?.availWidth;
    const height = window.screen?.availHeight;
    if (width === lastWidth && height === lastHeight) return;
    lastWidth = width;
    lastHeight = height;
    if (getUIScale() === UI_SCALE_AUTO) applyUIScale(UI_SCALE_AUTO);
  });
}
