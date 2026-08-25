export const UI_STYLES = ["modern", "classic"] as const;
export type UIStyle = (typeof UI_STYLES)[number];

const UI_STYLE_KEY = "orca-ui-style";

export function normalizeUIStyle(value: unknown): UIStyle {
  return value === "classic" ? "classic" : "modern";
}

export function getUIStyle(): UIStyle {
  try {
    return normalizeUIStyle(localStorage.getItem(UI_STYLE_KEY));
  } catch {
    return "modern";
  }
}

export function applyUIStyle(value: unknown, options: { persist?: boolean } = {}): UIStyle {
  const style = normalizeUIStyle(value);
  if (typeof document !== "undefined") {
    const root = document.documentElement;
    root.dataset.uiStyle = style;
    if (style === "modern") root.dataset.themeStyle = "slate";
    else root.removeAttribute("data-theme-style");
  }
  if (options.persist !== false) {
    try {
      localStorage.setItem(UI_STYLE_KEY, style);
    } catch {
      // Storage may be unavailable in isolated webviews.
    }
  }
  return style;
}

export function persistUIStyle(value: unknown): UIStyle {
	const style = normalizeUIStyle(value);
	try {
		localStorage.setItem(UI_STYLE_KEY, style);
	} catch {
		// Storage may be unavailable in isolated webviews.
	}
	return style;
}

export function initUIStyle(): void {
  applyUIStyle(getUIStyle(), { persist: false });
}
