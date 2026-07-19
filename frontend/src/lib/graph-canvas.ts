import type { Edge, Node } from "@xyflow/react";

/**
 * Data carried by canvas nodes. `nodeType` is the backend node-type id from
 * `GET /node-types`; `label`/`category` drive rendering in the custom node.
 */
export type CanvasNodeData = {
  label: string;
  nodeType: string;
  category?: string;
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
    },
  };
}
