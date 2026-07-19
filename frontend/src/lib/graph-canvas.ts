import type { Edge, Node } from "@xyflow/react";

/** Node data carried by canvas nodes. Custom node types arrive in #16. */
export type CanvasNodeData = {
  label: string;
};

export type CanvasNode = Node<CanvasNodeData>;
export type CanvasEdge = Edge;

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

/**
 * Build a new canvas node at the given position.
 *
 * Kept as a pure helper (no React Flow runtime) so node creation can be unit
 * tested without a DOM.
 */
export function createNode(
  position: { x: number; y: number },
  label?: string,
): CanvasNode {
  const id = nextNodeId();
  return {
    id,
    position,
    data: { label: label ?? `Node ${id.replace("node-", "")}` },
  };
}
