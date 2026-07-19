import { beforeEach, describe, expect, it } from "vitest";

import { createNode, nextNodeId, resetNodeSequence } from "./graph-canvas";

describe("graph-canvas helpers", () => {
  beforeEach(() => {
    resetNodeSequence(0);
  });

  it("generates unique, incrementing node ids", () => {
    expect(nextNodeId()).toBe("node-1");
    expect(nextNodeId()).toBe("node-2");
    expect(nextNodeId()).toBe("node-3");
  });

  it("creates a node at the given position with a default label", () => {
    const node = createNode({ x: 40, y: 80 });
    expect(node.id).toBe("node-1");
    expect(node.position).toEqual({ x: 40, y: 80 });
    expect(node.data.label).toBe("Node 1");
  });

  it("honours an explicit label", () => {
    const node = createNode({ x: 0, y: 0 }, "Start");
    expect(node.data.label).toBe("Start");
  });
});
