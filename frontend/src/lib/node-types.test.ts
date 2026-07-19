import { describe, expect, it } from "vitest";

import { groupByCategory, type NodeType } from "./node-types";

function makeType(id: string, category?: string): NodeType {
  return { id, name: id.toUpperCase(), category, configSchema: {} };
}

describe("groupByCategory", () => {
  it("groups node types by category, preserving order", () => {
    const groups = groupByCategory([
      makeType("start", "core"),
      makeType("http", "integration"),
      makeType("noop", "core"),
    ]);

    expect(groups).toEqual([
      ["core", [makeType("start", "core"), makeType("noop", "core")]],
      ["integration", [makeType("http", "integration")]],
    ]);
  });

  it("falls back to 'other' when category is missing or blank", () => {
    const groups = groupByCategory([makeType("a"), makeType("b", "  ")]);
    expect(groups).toEqual([["other", [makeType("a"), makeType("b", "  ")]]]);
  });

  it("returns an empty list for no node types", () => {
    expect(groupByCategory([])).toEqual([]);
  });
});
