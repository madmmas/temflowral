/** Viewport fitting helpers after async graph hydrate (#111). */

/** Padding passed to React Flow `fitView` after open/deep-link load. */
export const FIT_VIEW_PADDING = 0.2;

/** Animation duration (ms) for post-load fitView. */
export const FIT_VIEW_DURATION_MS = 200;

/**
 * Only auto-fit when the loaded graph has nodes. Empty graphs keep the
 * empty-state guide (#109) without a meaningless fitView call.
 */
export function shouldFitViewportAfterLoad(nodeCount: number): boolean {
  return Number.isFinite(nodeCount) && nodeCount > 0;
}
