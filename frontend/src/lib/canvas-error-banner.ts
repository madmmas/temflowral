/** Helpers for the canvas error banner (#108). */

export type CanvasBannerSource = "action" | "graphs" | null;

/**
 * Prefer the action/open/save/run error; fall back to graph-list failures.
 * Returns which source won so dismiss can clear the right state.
 */
export function pickCanvasBannerMessage(
  actionError: string | null | undefined,
  graphsError: string | null | undefined,
): { message: string; source: Exclude<CanvasBannerSource, null> } | null {
  const action = actionError?.trim();
  if (action) return { message: action, source: "action" };
  const graphs = graphsError?.trim();
  if (graphs) return { message: graphs, source: "graphs" };
  return null;
}

/** True when the banner was dismissed for this exact message+source. */
export function isCanvasBannerDismissed(
  dismissed: { source: Exclude<CanvasBannerSource, null>; message: string } | null,
  current: { source: Exclude<CanvasBannerSource, null>; message: string } | null,
): boolean {
  if (!dismissed || !current) return false;
  return (
    dismissed.source === current.source && dismissed.message === current.message
  );
}
