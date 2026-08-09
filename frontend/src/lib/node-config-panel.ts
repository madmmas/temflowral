/** Node config drawer width helpers (#110). */

export const NODE_CONFIG_PANEL_DEFAULT_WIDTH = 288; // w-72
export const NODE_CONFIG_PANEL_MIN_WIDTH = 240;
export const NODE_CONFIG_PANEL_MAX_WIDTH = 480;

export const NODE_CONFIG_WIDTH_STORAGE_KEY = "temflowral.nodeConfig.width";

/** Clamp a dragged config panel width to the allowed range. */
export function clampNodeConfigPanelWidth(width: number): number {
  if (!Number.isFinite(width)) return NODE_CONFIG_PANEL_DEFAULT_WIDTH;
  return Math.min(
    NODE_CONFIG_PANEL_MAX_WIDTH,
    Math.max(NODE_CONFIG_PANEL_MIN_WIDTH, Math.round(width)),
  );
}

export function readNodeConfigPanelWidth(
  storage: Pick<Storage, "getItem"> | null = defaultStorage(),
  fallback: number = NODE_CONFIG_PANEL_DEFAULT_WIDTH,
): number {
  if (!storage) return fallback;
  try {
    const raw = storage.getItem(NODE_CONFIG_WIDTH_STORAGE_KEY);
    if (raw === null) return fallback;
    const parsed = Number(raw);
    if (!Number.isFinite(parsed)) return fallback;
    return clampNodeConfigPanelWidth(parsed);
  } catch {
    return fallback;
  }
}

export function writeNodeConfigPanelWidth(
  width: number,
  storage: Pick<Storage, "setItem"> | null = defaultStorage(),
): void {
  if (!storage) return;
  try {
    storage.setItem(
      NODE_CONFIG_WIDTH_STORAGE_KEY,
      String(clampNodeConfigPanelWidth(width)),
    );
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
