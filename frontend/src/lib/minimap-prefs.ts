/** MiniMap visibility preference helpers (#104). */

export const MINIMAP_VISIBLE_STORAGE_KEY = "temflowral.minimap.visible";

/** Default visible when no preference is stored. */
export const MINIMAP_VISIBLE_DEFAULT = true;

export function readMinimapVisible(
  storage: Pick<Storage, "getItem"> | null = defaultStorage(),
  fallback: boolean = MINIMAP_VISIBLE_DEFAULT,
): boolean {
  if (!storage) return fallback;
  try {
    const raw = storage.getItem(MINIMAP_VISIBLE_STORAGE_KEY);
    if (raw === null) return fallback;
    if (raw === "true") return true;
    if (raw === "false") return false;
    return fallback;
  } catch {
    return fallback;
  }
}

export function writeMinimapVisible(
  visible: boolean,
  storage: Pick<Storage, "setItem"> | null = defaultStorage(),
): void {
  if (!storage) return;
  try {
    storage.setItem(MINIMAP_VISIBLE_STORAGE_KEY, String(visible));
  } catch {
    // Ignore private mode / quota errors.
  }
}

function defaultStorage(): Pick<Storage, "getItem" | "setItem"> | null {
  if (typeof window === "undefined") return null;
  try {
    return window.localStorage;
  } catch {
    return null;
  }
}

/** Compact MiniMap dimensions — less dominant than React Flow defaults. */
export const MINIMAP_STYLE = {
  width: 140,
  height: 92,
} as const;

/**
 * Dark-canvas MiniMap colors (avoids the default bright white panel).
 * Used whenever `prefers-color-scheme: dark` matches; light theme uses a
 * softer neutral panel instead of pure white.
 */
export const MINIMAP_COLORS_DARK = {
  bgColor: "#141414",
  maskColor: "rgba(0, 0, 0, 0.55)",
  nodeColor: "#525252",
  nodeStrokeColor: "#a3a3a3",
} as const;

export const MINIMAP_COLORS_LIGHT = {
  bgColor: "#f4f4f5",
  maskColor: "rgba(15, 15, 15, 0.12)",
  nodeColor: "#a1a1aa",
  nodeStrokeColor: "#52525b",
} as const;

export function minimapColorsForScheme(
  prefersDark: boolean,
): typeof MINIMAP_COLORS_DARK | typeof MINIMAP_COLORS_LIGHT {
  return prefersDark ? MINIMAP_COLORS_DARK : MINIMAP_COLORS_LIGHT;
}
