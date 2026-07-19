import { beforeEach, describe, expect, it } from "vitest";

import {
  createNode,
  nextNodeId,
  resetNodeSequence,
  WORKFLOW_NODE_TYPE,
} from "./graph-canvas";

describe("graph-canvas helpers", () => {
  beforeEach(() => {
    resetNodeSequence(0);
  });

  it("generates unique, incrementing node ids", () => {
    expect(nextNodeId()).toBe("node-1");
    expect(nextNodeId()).toBe("node-2");
    expect(nextNodeId()).toBe("node-3");
  });

  it("creates a typed node at the given position", () => {
    const node = createNode({ x: 40, y: 80 }, { nodeType: "http" });
    expect(node.id).toBe("node-1");
    expect(node.type).toBe(WORKFLOW_NODE_TYPE);
    expect(node.position).toEqual({ x: 40, y: 80 });
    expect(node.data).toEqual({
      label: "http",
      nodeType: "http",
      category: undefined,
    });
  });

  it("honours an explicit label and category", () => {
    const node = createNode(
      { x: 0, y: 0 },
      { nodeType: "http", label: "Fetch data", category: "integration" },
    );
    expect(node.data.label).toBe("Fetch data");
    expect(node.data.category).toBe("integration");
  });
});
