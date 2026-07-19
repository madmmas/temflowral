import { beforeEach, describe, expect, it } from "vitest";

import {
  apiErrorMessage,
  createNode,
  isTerminalRunStatus,
  nextNodeId,
  resetNodeSequence,
  serializeGraph,
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
      config: {},
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

  it("serializes canvas nodes and edges to the contract request", () => {
    const start = createNode(
      { x: 10, y: 20 },
      { nodeType: "start", label: "Start" },
    );
    const noop = createNode(
      { x: 100, y: 120 },
      { nodeType: "noop", label: "No-op", config: { value: "hello" } },
    );

    expect(
      serializeGraph("  Demo graph  ", [start, noop], [
        {
          id: "edge-1",
          source: start.id,
          target: noop.id,
          sourceHandle: null,
          targetHandle: null,
        },
      ]),
    ).toEqual({
      name: "Demo graph",
      nodes: [
        {
          id: "node-1",
          type: "start",
          label: "Start",
          position: { x: 10, y: 20 },
          config: {},
        },
        {
          id: "node-2",
          type: "noop",
          label: "No-op",
          position: { x: 100, y: 120 },
          config: { value: "hello" },
        },
      ],
      edges: [
        {
          id: "edge-1",
          source: "node-1",
          target: "node-2",
          sourceHandle: undefined,
          targetHandle: undefined,
        },
      ],
    });
  });

  it("omits a blank graph name", () => {
    expect(serializeGraph("  ", [], []).name).toBeUndefined();
  });

  it("identifies terminal run statuses", () => {
    expect(isTerminalRunStatus("pending")).toBe(false);
    expect(isTerminalRunStatus("running")).toBe(false);
    expect(isTerminalRunStatus("completed")).toBe(true);
    expect(isTerminalRunStatus("failed")).toBe(true);
    expect(isTerminalRunStatus("cancelled")).toBe(true);
  });

  it("extracts contract error messages with a fallback", () => {
    expect(apiErrorMessage({ message: "graph is invalid" }, "fallback")).toBe(
      "graph is invalid",
    );
    expect(apiErrorMessage(null, "fallback")).toBe("fallback");
  });
});
