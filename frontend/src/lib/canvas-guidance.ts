/** Empty-canvas guidance + authoring tip prefs (#109). */

export const AUTHORING_TIP_DISMISSED_STORAGE_KEY =
  "temflowral.authoringTip.dismissed";

export const EMPTY_CANVAS_HEADING = "Build a workflow";
export const EMPTY_CANVAS_BODY =
  "Add Start from the palette, connect nodes, then Save and Run.";

export const AUTHORING_TIP_TEXT =
  "Select a node to edit config · drag handles to connect · Delete to remove";

/** Tip is visible by default until the user dismisses it. */
export function readAuthoringTipDismissed(
  storage: Pick<Storage, "getItem"> | null = defaultStorage(),
): boolean {
  if (!storage) return false;
  try {
    return storage.getItem(AUTHORING_TIP_DISMISSED_STORAGE_KEY) === "true";
  } catch {
    return false;
  }
}

export function writeAuthoringTipDismissed(
  dismissed: boolean,
  storage: Pick<Storage, "setItem" | "removeItem"> | null = defaultStorage(),
): void {
  if (!storage) return;
  try {
    if (dismissed) {
      storage.setItem(AUTHORING_TIP_DISMISSED_STORAGE_KEY, "true");
    } else {
      storage.removeItem(AUTHORING_TIP_DISMISSED_STORAGE_KEY);
    }
  } catch {
    // Ignore private mode / quota errors.
  }
}

/**
 * Slight screen-space offset for palette clicks so repeated center drops
 * do not stack perfectly on top of each other.
 */
export function paletteClickScreenOffset(nodeCount: number): {
  x: number;
  y: number;
} {
  const safe = Math.max(0, Math.floor(nodeCount));
  const step = 32;
  const col = safe % 4;
  const row = Math.floor(safe / 4) % 4;
  return { x: col * step, y: row * step };
}

function defaultStorage(): Pick<
  Storage,
  "getItem" | "setItem" | "removeItem"
> | null {
  if (typeof window === "undefined") return null;
  try {
    return window.localStorage;
  } catch {
    return null;
  }
}
