import type { components } from "@/api";
import type { Edge, Node } from "@xyflow/react";

/**
 * Data carried by canvas nodes. `nodeType` is the backend node-type id from
 * `GET /node-types`; `label`/`category` drive rendering in the custom node.
 */
export type CanvasNodeData = {
  label: string;
  nodeType: string;
  category?: string;
  config?: Record<string, unknown>;
};

export type CanvasNode = Node<CanvasNodeData>;
export type CanvasEdge = Edge;

/** React Flow node.type used for every palette-created node (custom renderer). */
export const WORKFLOW_NODE_TYPE = "workflow";

let nodeSequence = 0;

/** Reset the id counter. Intended for tests. */
export function resetNodeSequence(value = 0): void {
  nodeSequence = value;
}

/** Generate a stable, unique node id. */
export function nextNodeId(): string {
  nodeSequence += 1;
  return `node-${nodeSequence}`;
}

export type NewNodeInput = {
  nodeType: string;
  label?: string;
  category?: string;
  config?: Record<string, unknown>;
};

/**
 * Build a new canvas node of the given node type at a position.
 *
 * Pure helper (no React Flow runtime) so node creation is unit-testable
 * without a DOM.
 */
export function createNode(
  position: { x: number; y: number },
  input: NewNodeInput,
): CanvasNode {
  const id = nextNodeId();
  return {
    id,
    type: WORKFLOW_NODE_TYPE,
    position,
    data: {
      label: input.label ?? input.nodeType,
      nodeType: input.nodeType,
      category: input.category,
      config: input.config ?? {},
    },
  };
}

export type CreateGraphRequest = components["schemas"]["CreateGraphRequest"];
export type UpdateGraphRequest = components["schemas"]["UpdateGraphRequest"];
export type Graph = components["schemas"]["Graph"];
export type GraphSummary = components["schemas"]["GraphSummary"];

/** Convert React Flow state to the OpenAPI create/update graph request shape. */
export function serializeGraph(
  name: string,
  nodes: readonly CanvasNode[],
  edges: readonly CanvasEdge[],
): CreateGraphRequest & UpdateGraphRequest {
  return {
    name: name.trim() || undefined,
    nodes: nodes.map((node) => ({
      id: node.id,
      type: node.data.nodeType,
      label: node.data.label,
      position: {
        x: node.position.x,
        y: node.position.y,
      },
      config: node.data.config ?? {},
    })),
    edges: edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceHandle: edge.sourceHandle ?? undefined,
      targetHandle: edge.targetHandle ?? undefined,
    })),
  };
}

/**
 * Convert a persisted Graph into React Flow nodes/edges and a display name.
 * Advances the local node id sequence past any `node-N` ids already present.
 */
export function deserializeGraph(graph: Graph): {
  name: string;
  nodes: CanvasNode[];
  edges: CanvasEdge[];
} {
  const nodes: CanvasNode[] = graph.nodes.map((node) => ({
    id: node.id,
    type: WORKFLOW_NODE_TYPE,
    position: { x: node.position.x, y: node.position.y },
    data: {
      label: node.label ?? node.type,
      nodeType: node.type,
      config: (node.config as Record<string, unknown> | undefined) ?? {},
    },
  }));
  syncNodeSequenceFromIds(nodes.map((node) => node.id));

  return {
    name: graph.name?.trim() || "Untitled workflow",
    nodes,
    edges: graph.edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceHandle: edge.sourceHandle ?? undefined,
      targetHandle: edge.targetHandle ?? undefined,
    })),
  };
}

/** Keep `nextNodeId()` from colliding with loaded `node-N` ids. */
export function syncNodeSequenceFromIds(ids: readonly string[]): void {
  let max = nodeSequence;
  for (const id of ids) {
    const match = /^node-(\d+)$/.exec(id);
    if (!match) continue;
    const value = Number(match[1]);
    if (Number.isFinite(value) && value > max) {
      max = value;
    }
  }
  nodeSequence = max;
}

export type RunStatus = components["schemas"]["RunStatus"];

export function isTerminalRunStatus(status: RunStatus): boolean {
  return status === "completed" || status === "failed" || status === "cancelled";
}

/** Read a contract Error.message when available, otherwise return the fallback. */
export function apiErrorMessage(error: unknown, fallback: string): string {
  if (
    typeof error === "object" &&
    error !== null &&
    "message" in error &&
    typeof error.message === "string"
  ) {
    return error.message;
  }
  return fallback;
}
