// theme.ts keeps the desktop shell on the fixed O.R.C.A light product
// theme. Older versions exposed multiple modes/styles, so the helpers still
// accept legacy inputs but normalize them to the single supported appearance.
//
// When running inside the Wails shell, applyTheme also syncs the native window
// theme (title bar, traffic lights, etc.) so the OS chrome matches the webview.

import {
  WindowSetLightTheme,
  WindowSetBackgroundColour,
} from "../../wailsjs/runtime/runtime";

export type Theme = "auto" | "light" | "dark";
export type ResolvedTheme = Exclude<Theme, "auto">;

export const THEME_STYLES = ["slate"] as const;

export type ThemeStyle = (typeof THEME_STYLES)[number];

// Old style identifiers resolve to the fixed default so stale local settings
// cannot switch the app into another visual direction.
const LEGACY_STYLE_MAP: Record<string, ThemeStyle> = {
  graphite: "slate",
  aurora: "slate",
  carbon: "slate",
  nocturne: "slate",
  amber: "slate",
  ember: "slate",
  midnight: "slate",
  sandstone: "slate",
  porcelain: "slate",
  linen: "slate",
  glacier: "slate",
  pop: "slate",
  poppaint: "slate",
  "pop-paint": "slate",
};

const DEFAULT_THEME_STYLE: ThemeStyle = "slate";
const DEFAULT_THEME: Theme = "light";

const THEME_KEY = "orca-theme";
const STYLE_KEY = "orca-theme-style";
let currentTheme: Theme = DEFAULT_THEME;
let currentThemeStyle: ThemeStyle = DEFAULT_THEME_STYLE;

export function normalizeThemePreference(value: unknown): Theme {
  if (typeof value === "object" && value !== null) {
    return normalizeThemePreference((value as { mode?: unknown }).mode);
  }
  void value;
  return DEFAULT_THEME;
}

export function isThemeStyle(value: unknown): value is ThemeStyle {
  return typeof value === "string" && (THEME_STYLES as readonly string[]).includes(value);
}

export function getTheme(): Theme {
  return currentTheme;
}

export function getResolvedTheme(theme: Theme = getTheme()): ResolvedTheme {
  void theme;
  return "light";
}

// Direction is orthogonal to theme, but keep this helper so callers that
// stored values in the old "style implies theme" model can still ask.
export function defaultStyleForTheme(_theme: Theme = getTheme()): ThemeStyle {
  return DEFAULT_THEME_STYLE;
}

// themeForStyle previously returned the dark/light forced by the style. Style
// is now independent of theme, so we keep the current theme.
export function themeForStyle(_style: ThemeStyle): ResolvedTheme {
  return getResolvedTheme();
}

export function getThemeStyle(_theme: Theme = getTheme()): ThemeStyle {
  return currentThemeStyle;
}

export function normalizeThemeStyleForTheme(style: string | undefined, _theme?: Theme): ThemeStyle {
  if (typeof style !== "string") return DEFAULT_THEME_STYLE;
  if (isThemeStyle(style)) return style;
  return LEGACY_STYLE_MAP[style] ?? DEFAULT_THEME_STYLE;
}

export function applyTheme(theme: Theme, style: ThemeStyle = getThemeStyle(theme), options: { persist?: boolean } = {}): void {
  if (typeof document === "undefined") return;
  const root = document.documentElement;
  root.removeAttribute("data-theme-mode");
  root.removeAttribute("data-theme-scheme");
  root.setAttribute("data-theme", DEFAULT_THEME);

  void theme;
  void style;
  currentTheme = DEFAULT_THEME;
  currentThemeStyle = DEFAULT_THEME_STYLE;
  root.setAttribute("data-theme-style", DEFAULT_THEME_STYLE);

  // Sync the native window theme (title bar, traffic lights) to match.
  if (typeof window !== "undefined" && window.runtime) {
    WindowSetLightTheme();
  }

  void options;
}

export function readLegacyThemePreference(): { theme: Theme; style: ThemeStyle; hasValue: boolean } {
  if (typeof localStorage === "undefined") return { theme: DEFAULT_THEME, style: DEFAULT_THEME_STYLE, hasValue: false };
  let rawTheme: string | null = null;
  let rawStyle: string | null = null;
  try {
    rawTheme = localStorage.getItem(THEME_KEY);
    rawStyle = localStorage.getItem(STYLE_KEY);
  } catch {
    return { theme: DEFAULT_THEME, style: DEFAULT_THEME_STYLE, hasValue: false };
  }
  const hasValue = rawTheme !== null || rawStyle !== null;
  return { theme: DEFAULT_THEME, style: DEFAULT_THEME_STYLE, hasValue };
}

export function clearLegacyThemePreference(): void {
  try {
    localStorage.removeItem(THEME_KEY);
    localStorage.removeItem(STYLE_KEY);
  } catch {
    /* ignore storage failures */
  }
}

// initTheme runs before React mounts. It applies the saved theme to the DOM and
// sets the native window background colour to match the resolved theme, avoiding
// a white (or wrong-colour) flash while the webview paints its first frame.
export function initTheme(): void {
  const theme = getTheme();
  applyTheme(theme, getThemeStyle(theme), { persist: false });

  if (typeof window !== "undefined" && window.runtime) {
    // Light shell: matches the O.R.C.A blue-white surface.
    WindowSetBackgroundColour(246, 250, 255, 255);
  }
}
