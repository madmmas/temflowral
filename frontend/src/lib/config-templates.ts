import type { CanvasEdge, CanvasNode } from "@/lib/graph-canvas";

/** Common output paths by node type for template assist (#112). */
const OUTPUT_PATHS_BY_TYPE: Record<string, string[]> = {
  start: [""],
  noop: [""],
  http: ["", "statusCode", "body"],
  delay: ["", "seconds"],
  condition: ["", "matched", "branch"],
  wait: ["", "signal", "timedOut", "branch", "payload"],
  childWorkflow: ["", "type", "nodes"],
};

/**
 * Transitive upstream node ids (predecessors) for `nodeId`, excluding self.
 * Order is BFS from direct parents so nearer nodes appear first.
 */
export function upstreamNodeIds(
  edges: readonly CanvasEdge[],
  nodeId: string,
): string[] {
  const parents = new Map<string, string[]>();
  for (const edge of edges) {
    const list = parents.get(edge.target) ?? [];
    list.push(edge.source);
    parents.set(edge.target, list);
  }

  const ordered: string[] = [];
  const seen = new Set<string>([nodeId]);
  const queue = [...(parents.get(nodeId) ?? [])];
  while (queue.length > 0) {
    const id = queue.shift()!;
    if (seen.has(id)) continue;
    seen.add(id);
    ordered.push(id);
    for (const parent of parents.get(id) ?? []) {
      if (!seen.has(parent)) queue.push(parent);
    }
  }
  return ordered;
}

export function formatTemplateSuggestion(
  nodeId: string,
  outputPath: string,
): string {
  if (!outputPath) return `{{ nodes.${nodeId}.output }}`;
  return `{{ nodes.${nodeId}.output.${outputPath} }}`;
}

/**
 * Template insert suggestions for the selected node's string fields.
 * Includes bare `{{` completions for upstream nodes and known output paths.
 */
export function templateSuggestionsForNode(
  nodes: readonly CanvasNode[],
  edges: readonly CanvasEdge[],
  selectedNodeId: string,
): string[] {
  const byId = new Map(nodes.map((node) => [node.id, node]));
  const suggestions: string[] = [];
  const seen = new Set<string>();

  for (const upstreamId of upstreamNodeIds(edges, selectedNodeId)) {
    const node = byId.get(upstreamId);
    const paths = OUTPUT_PATHS_BY_TYPE[node?.data.nodeType ?? ""] ?? [""];
    for (const path of paths) {
      const suggestion = formatTemplateSuggestion(upstreamId, path);
      if (seen.has(suggestion)) continue;
      seen.add(suggestion);
      suggestions.push(suggestion);
    }
  }
  return suggestions;
}

/** Filter suggestions by the unfinished `{{...` fragment at the caret. */
export function filterTemplateSuggestions(
  suggestions: readonly string[],
  fieldValue: string,
  caretIndex: number,
): string[] {
  const before = fieldValue.slice(0, Math.max(0, caretIndex));
  const open = before.lastIndexOf("{{");
  if (open < 0) return [];
  const afterOpen = before.slice(open);
  if (afterOpen.includes("}}")) return [];
  const needle = afterOpen.toLowerCase();
  return suggestions.filter((entry) => entry.toLowerCase().startsWith(needle));
}
