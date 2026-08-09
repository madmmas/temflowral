/** Pure helpers for the run result drawer (#102). */

export const RUN_RESULT_PANEL_DEFAULT_HEIGHT = 200;
export const RUN_RESULT_PANEL_MIN_HEIGHT = 96;
export const RUN_RESULT_PANEL_MAX_HEIGHT = 480;
export const RUN_RESULT_PANEL_COLLAPSED_HEIGHT = 36;

/** Pretty-print run.result / structured payloads for the drawer. */
export function formatRunPayload(payload: unknown): string {
  if (typeof payload === "string") return payload;
  try {
    return JSON.stringify(payload, null, 2);
  } catch {
    return String(payload);
  }
}

/** Clamp a dragged panel height to the allowed range. */
export function clampRunResultPanelHeight(height: number): number {
  if (!Number.isFinite(height)) return RUN_RESULT_PANEL_DEFAULT_HEIGHT;
  return Math.min(
    RUN_RESULT_PANEL_MAX_HEIGHT,
    Math.max(RUN_RESULT_PANEL_MIN_HEIGHT, Math.round(height)),
  );
}

/**
 * Whether the drawer should offer content for this run.
 * Result and/or error belong in the panel (not the footer strip).
 */
export function runHasResultPanelContent(run: {
  result?: unknown;
  error?: string | null;
} | null): boolean {
  if (!run) return false;
  if (run.error != null && run.error !== "") return true;
  return run.result !== undefined && run.result !== null;
}
