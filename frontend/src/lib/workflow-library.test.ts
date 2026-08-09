import { describe, expect, it } from "vitest";

import type { GraphSummary } from "@/lib/graph-canvas";
import {
  filterAndSortGraphs,
  formatGraphUpdatedAt,
  graphDisplayName,
  graphShortId,
} from "@/lib/workflow-library";

function summary(
  partial: Partial<GraphSummary> & Pick<GraphSummary, "id">,
): GraphSummary {
  return {
    createdAt: "2026-08-01T00:00:00Z",
    updatedAt: "2026-08-01T00:00:00Z",
    ...partial,
  };
}

describe("graphDisplayName", () => {
  it("falls back to Untitled when name missing or blank", () => {
    expect(graphDisplayName(summary({ id: "a" }))).toBe("Untitled");
    expect(graphDisplayName(summary({ id: "a", name: "  " }))).toBe("Untitled");
    expect(graphDisplayName(summary({ id: "a", name: "Demo" }))).toBe("Demo");
  });
});

describe("graphShortId", () => {
  it("returns the first 8 characters", () => {
    expect(graphShortId("41c5aff8-349e-417c-a16a-1a8d2c2e37cb")).toBe(
      "41c5aff8",
    );
  });
});

describe("formatGraphUpdatedAt", () => {
  it("formats a valid timestamp", () => {
    const formatted = formatGraphUpdatedAt(
      "2026-08-08T11:57:13.019494Z",
      new Date("2026-08-09T00:00:00Z"),
    );
    expect(formatted.length).toBeGreaterThan(4);
    expect(formatted).toMatch(/8/);
  });

  it("returns the raw string when invalid", () => {
    expect(formatGraphUpdatedAt("not-a-date")).toBe("not-a-date");
  });
});

describe("filterAndSortGraphs", () => {
  const graphs = [
    summary({
      id: "aaaaaaa1-0000-4000-8000-000000000001",
      name: "Happy path workflow",
      updatedAt: "2026-08-08T10:00:00Z",
    }),
    summary({
      id: "bbbbbbb2-0000-4000-8000-000000000002",
      name: "Untitled workflow",
      updatedAt: "2026-08-08T12:00:00Z",
    }),
    summary({
      id: "ccccccc3-0000-4000-8000-000000000003",
      name: "Happy path workflow",
      updatedAt: "2026-08-08T11:00:00Z",
    }),
  ];

  it("sorts by updatedAt descending by default", () => {
    expect(filterAndSortGraphs(graphs, "").map((g) => g.id.slice(0, 8))).toEqual(
      ["bbbbbbb2", "ccccccc3", "aaaaaaa1"],
    );
  });

  it("sorts by name then updatedAt", () => {
    const sorted = filterAndSortGraphs(graphs, "", "name");
    expect(sorted.map((g) => [graphDisplayName(g), g.id.slice(0, 8)])).toEqual([
      ["Happy path workflow", "ccccccc3"],
      ["Happy path workflow", "aaaaaaa1"],
      ["Untitled workflow", "bbbbbbb2"],
    ]);
  });

  it("filters by name substring", () => {
    expect(filterAndSortGraphs(graphs, "untitled")).toHaveLength(1);
    expect(filterAndSortGraphs(graphs, "HAPPY")).toHaveLength(2);
  });

  it("filters by id substring", () => {
    expect(filterAndSortGraphs(graphs, "bbbbbbb2")).toHaveLength(1);
  });
});
