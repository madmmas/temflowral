import { beforeEach, describe, expect, it } from "vitest";

import {
  apiErrorMessage,
  createNode,
  deserializeGraph,
  graphFingerprint,
  isTerminalRunStatus,
  nextNodeId,
  resetNodeSequence,
  serializeGraph,
  syncNodeSequenceFromIds,
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

  it("deserializes a persisted graph back onto the canvas", () => {
    const loaded = deserializeGraph({
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "Saved workflow",
      nodes: [
        {
          id: "start-1",
          type: "start",
          label: "Start",
          position: { x: 1, y: 2 },
          config: {},
        },
        {
          id: "node-7",
          type: "noop",
          position: { x: 3, y: 4 },
          config: { value: 1 },
        },
      ],
      edges: [{ id: "e1", source: "start-1", target: "node-7" }],
      createdAt: "2026-08-08T00:00:00Z",
      updatedAt: "2026-08-08T01:00:00Z",
    });

    expect(loaded.name).toBe("Saved workflow");
    expect(loaded.nodes).toHaveLength(2);
    expect(loaded.nodes[0]).toMatchObject({
      id: "start-1",
      type: WORKFLOW_NODE_TYPE,
      data: { label: "Start", nodeType: "start", config: {} },
    });
    expect(loaded.edges).toEqual([
      {
        id: "e1",
        source: "start-1",
        target: "node-7",
        sourceHandle: undefined,
        targetHandle: undefined,
      },
    ]);
    expect(nextNodeId()).toBe("node-8");
  });

  it("syncs the node sequence from loaded ids", () => {
    syncNodeSequenceFromIds(["start-1", "node-4", "noop-2"]);
    expect(nextNodeId()).toBe("node-5");
  });

  it("round-trips named sourceHandle on edges", () => {
    const condition = createNode(
      { x: 0, y: 0 },
      { nodeType: "condition", label: "Branch" },
    );
    const yes = createNode({ x: 100, y: 0 }, { nodeType: "noop", label: "Yes" });
    const no = createNode({ x: 100, y: 80 }, { nodeType: "noop", label: "No" });

    const serialized = serializeGraph("Branches", [condition, yes, no], [
      {
        id: "e-true",
        source: condition.id,
        target: yes.id,
        sourceHandle: "true",
      },
      {
        id: "e-false",
        source: condition.id,
        target: no.id,
        sourceHandle: "false",
      },
    ]);

    expect(serialized.edges).toEqual([
      {
        id: "e-true",
        source: "node-1",
        target: "node-2",
        sourceHandle: "true",
        targetHandle: undefined,
      },
      {
        id: "e-false",
        source: "node-1",
        target: "node-3",
        sourceHandle: "false",
        targetHandle: undefined,
      },
    ]);

    const loaded = deserializeGraph({
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "Branches",
      nodes: serialized.nodes!.map((node) => ({
        ...node,
        position: node.position ?? { x: 0, y: 0 },
        config: node.config ?? {},
      })),
      edges: serialized.edges!,
      createdAt: "2026-08-08T00:00:00Z",
      updatedAt: "2026-08-08T00:00:00Z",
    });

    expect(loaded.edges.map((edge) => edge.sourceHandle)).toEqual([
      "true",
      "false",
    ]);
  });

  it("serializes activityOptions and taskQueue", () => {
    const noop = createNode(
      { x: 0, y: 0 },
      {
        nodeType: "noop",
        label: "No-op",
        activityOptions: {
          startToCloseTimeoutSeconds: 60,
          retryPolicy: { maximumAttempts: 2 },
        },
        taskQueue: "worker.gpu",
      },
    );
    expect(serializeGraph("Opts", [noop], []).nodes?.[0]).toMatchObject({
      activityOptions: {
        startToCloseTimeoutSeconds: 60,
        retryPolicy: { maximumAttempts: 2 },
      },
      taskQueue: "worker.gpu",
    });
  });

  it("deserializes activityOptions and taskQueue", () => {
    const loaded = deserializeGraph({
      id: "550e8400-e29b-41d4-a716-446655440000",
      name: "Opts",
      nodes: [
        {
          id: "noop-1",
          type: "noop",
          position: { x: 0, y: 0 },
          config: {},
          activityOptions: { startToCloseTimeoutSeconds: 30 },
          taskQueue: "special",
        },
      ],
      edges: [],
      createdAt: "2026-08-08T00:00:00Z",
      updatedAt: "2026-08-08T00:00:00Z",
    });
    expect(loaded.nodes[0].data.activityOptions).toEqual({
      startToCloseTimeoutSeconds: 30,
    });
    expect(loaded.nodes[0].data.taskQueue).toBe("special");
  });

  it("fingerprints authoring state for dirty detection", () => {
    const node = createNode({ x: 1, y: 2 }, { nodeType: "start", label: "S" });
    const a = graphFingerprint("A", [node], []);
    const b = graphFingerprint("B", [node], []);
    expect(a).not.toBe(b);
    expect(graphFingerprint("A", [node], [])).toBe(a);
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
    expect(
      apiErrorMessage({ error: { message: "nested failure" } }, "fallback"),
    ).toBe("nested failure");
    expect(apiErrorMessage(null, "fallback")).toBe("fallback");
  });
});
