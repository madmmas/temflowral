import { describe, expect, it } from "vitest";

import type { CanvasEdge, CanvasNode } from "@/lib/graph-canvas";
import {
  filterTemplateSuggestions,
  formatTemplateSuggestion,
  templateSuggestionsForNode,
  upstreamNodeIds,
} from "@/lib/config-templates";

function node(
  id: string,
  nodeType: string,
): CanvasNode {
  return {
    id,
    type: "workflow",
    position: { x: 0, y: 0 },
    data: { label: id, nodeType, config: {} },
  };
}

describe("upstreamNodeIds", () => {
  it("walks transitive predecessors", () => {
    const edges: CanvasEdge[] = [
      { id: "e1", source: "a", target: "b" },
      { id: "e2", source: "b", target: "c" },
    ];
    expect(upstreamNodeIds(edges, "c")).toEqual(["b", "a"]);
    expect(upstreamNodeIds(edges, "a")).toEqual([]);
  });
});

describe("templateSuggestionsForNode", () => {
  it("suggests upstream http outputs", () => {
    const nodes = [node("start-1", "start"), node("http-1", "http"), node("n", "noop")];
    const edges: CanvasEdge[] = [
      { id: "e1", source: "start-1", target: "http-1" },
      { id: "e2", source: "http-1", target: "n" },
    ];
    const suggestions = templateSuggestionsForNode(nodes, edges, "n");
    expect(suggestions).toContain(formatTemplateSuggestion("http-1", "statusCode"));
    expect(suggestions).toContain(formatTemplateSuggestion("start-1", ""));
    expect(suggestions.some((s) => s.includes("n.output"))).toBe(false);
  });
});

describe("filterTemplateSuggestions", () => {
  it("filters by the open {{ fragment", () => {
    const all = [
      "{{ nodes.a.output }}",
      "{{ nodes.a.output.statusCode }}",
      "{{ nodes.b.output }}",
    ];
    expect(
      filterTemplateSuggestions(all, "x {{ nodes.a", "{{ nodes.a".length + 2),
    ).toEqual([
      "{{ nodes.a.output }}",
      "{{ nodes.a.output.statusCode }}",
    ]);
    expect(filterTemplateSuggestions(all, "plain", 5)).toEqual([]);
  });
});
