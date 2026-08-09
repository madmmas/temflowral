/**
 * Shared accessibility helpers for the reference canvas (#114).
 * Baseline only — not a WCAG claim; see docs/canvas-accessibility.md.
 */

/** True when keyboard delete/backspace should not remove canvas nodes. */
export function isTypingTarget(target: EventTarget | null): boolean {
  if (!target || typeof target !== "object") return false;
  const el = target as {
    tagName?: unknown;
    isContentEditable?: unknown;
  };
  if (typeof el.tagName !== "string") return false;
  const tag = el.tagName.toUpperCase();
  if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") {
    return true;
  }
  return el.isContentEditable === true;
}

export type EscapeLayer = "library" | "config" | "result" | "none";

/** Prefer closing the topmost overlay first (library → config → result). */
export function resolveEscapeLayer(state: {
  libraryOpen: boolean;
  configOpen: boolean;
  resultOpen: boolean;
}): EscapeLayer {
  if (state.libraryOpen) return "library";
  if (state.configOpen) return "config";
  if (state.resultOpen) return "result";
  return "none";
}
