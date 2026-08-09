import type { GraphSummary } from "@/lib/graph-canvas";

/** Default canvas name for brand-new graphs (#107). */
export const DEFAULT_GRAPH_NAME = "Untitled workflow";

export type NamePrompt = (
  message: string,
  defaultValue?: string,
) => string | null;
export type NameConfirm = (message: string) => boolean;

export type ResolveGraphNameResult =
  | { ok: true; name: string }
  | {
      ok: false;
      reason: "cancelled" | "placeholder" | "duplicate_declined";
      message?: string;
    };

/** Trim; blank becomes the default Untitled label. */
export function normalizeGraphName(name: string): string {
  const trimmed = name.trim();
  return trimmed || DEFAULT_GRAPH_NAME;
}

/**
 * True when the name is blank or the default Untitled label
 * (case-insensitive). First Save should nudge away from this.
 */
export function isPlaceholderGraphName(name: string): boolean {
  const trimmed = name.trim();
  if (!trimmed) return true;
  return trimmed.toLowerCase() === DEFAULT_GRAPH_NAME.toLowerCase();
}

/** Case-insensitive name match against other saved graphs. */
export function findGraphsWithSameName(
  graphs: readonly GraphSummary[],
  name: string,
  excludeGraphId?: string | null,
): GraphSummary[] {
  const needle = name.trim().toLowerCase();
  if (!needle) return [];
  return graphs.filter((graph) => {
    if (excludeGraphId && graph.id === excludeGraphId) return false;
    const other = graph.name?.trim().toLowerCase() ?? "";
    return other === needle;
  });
}

export function duplicateNameConfirmMessage(
  name: string,
  duplicateCount: number,
): string {
  const label = name.trim() || DEFAULT_GRAPH_NAME;
  if (duplicateCount <= 1) {
    return `A workflow named "${label}" already exists. Save anyway?`;
  }
  return `${duplicateCount} workflows are already named "${label}". Save anyway?`;
}

/**
 * Resolve the name used for Save/Run.
 * - New graphs with a placeholder name: prompt for a real name.
 * - Duplicate names: soft confirm (API still allows non-unique names).
 */
export function resolveGraphNameForSave(options: {
  name: string;
  isNewGraph: boolean;
  graphs: readonly GraphSummary[];
  excludeGraphId?: string | null;
  prompt: NamePrompt;
  confirm: NameConfirm;
}): ResolveGraphNameResult {
  let name = options.name.trim();

  if (options.isNewGraph && isPlaceholderGraphName(name)) {
    const prompted = options.prompt(
      "Name this workflow before saving:",
      "",
    );
    if (prompted === null) {
      return { ok: false, reason: "cancelled" };
    }
    name = prompted.trim();
    if (isPlaceholderGraphName(name)) {
      return {
        ok: false,
        reason: "placeholder",
        message:
          "Choose a name other than the default before saving a new workflow.",
      };
    }
  } else {
    name = normalizeGraphName(options.name);
  }

  const duplicates = findGraphsWithSameName(
    options.graphs,
    name,
    options.excludeGraphId,
  );
  if (duplicates.length > 0) {
    const accepted = options.confirm(
      duplicateNameConfirmMessage(name, duplicates.length),
    );
    if (!accepted) {
      return { ok: false, reason: "duplicate_declined" };
    }
  }

  return { ok: true, name };
}

/** Inline toolbar hint for placeholder / duplicate names. */
export function graphNameHint(options: {
  name: string;
  isNewGraph: boolean;
  duplicateCount: number;
}): string | null {
  if (options.isNewGraph && isPlaceholderGraphName(options.name)) {
    return "Name this workflow before first save";
  }
  if (options.duplicateCount > 0 && options.name.trim()) {
    return options.duplicateCount === 1
      ? "Another saved workflow uses this name"
      : `${options.duplicateCount} saved workflows use this name`;
  }
  return null;
}
