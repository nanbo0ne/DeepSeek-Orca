export const UI_SCALE_AUTO = 0;
export const UI_SCALE_MIN = 80;
export const UI_SCALE_MAX = 125;
export const UI_SCALE_STEP = 5;

const UI_SCALE_KEY = "deepseek-orca-ui-scale";

export function isUIScale(value: unknown): value is number {
  return typeof value === "number" && Number.isInteger(value) && (value === UI_SCALE_AUTO || (value >= UI_SCALE_MIN && value <= UI_SCALE_MAX && value % UI_SCALE_STEP === 0));
}

export function effectiveUIScale(preference: number): number {
  // Wails is PerMonitorV2-aware, so Windows has already applied the current
  // monitor's DPI. Automatic mode must not add a second resolution-based zoom.
  return preference === UI_SCALE_AUTO ? 100 : isUIScale(preference) ? preference : 100;
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
}
