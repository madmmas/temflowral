import type { GraphSummary } from "@/lib/graph-canvas";

export type WorkflowLibrarySort = "updatedAt" | "name";

export function graphDisplayName(graph: GraphSummary): string {
  const name = graph.name?.trim();
  return name || "Untitled";
}

export function graphShortId(graphId: string): string {
  return graphId.slice(0, 8);
}

/** Format updatedAt for the library list (local timezone). */
export function formatGraphUpdatedAt(
  updatedAt: string,
  now: Date = new Date(),
): string {
  const date = new Date(updatedAt);
  if (Number.isNaN(date.getTime())) return updatedAt;

  const sameYear = date.getFullYear() === now.getFullYear();
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    year: sameYear ? undefined : "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}

/**
 * Filter by name or id (case-insensitive substring), then sort.
 * Default sort: newest `updatedAt` first.
 */
export function filterAndSortGraphs(
  graphs: readonly GraphSummary[],
  query: string,
  sort: WorkflowLibrarySort = "updatedAt",
): GraphSummary[] {
  const needle = query.trim().toLowerCase();
  const filtered = needle
    ? graphs.filter((graph) => {
        const name = graphDisplayName(graph).toLowerCase();
        const id = graph.id.toLowerCase();
        return name.includes(needle) || id.includes(needle);
      })
    : [...graphs];

  filtered.sort((a, b) => {
    if (sort === "name") {
      const byName = graphDisplayName(a).localeCompare(
        graphDisplayName(b),
        undefined,
        { sensitivity: "base" },
      );
      if (byName !== 0) return byName;
      return b.updatedAt.localeCompare(a.updatedAt);
    }
    const byUpdated = b.updatedAt.localeCompare(a.updatedAt);
    if (byUpdated !== 0) return byUpdated;
    return graphDisplayName(a).localeCompare(graphDisplayName(b), undefined, {
      sensitivity: "base",
    });
  });

  return filtered;
}
