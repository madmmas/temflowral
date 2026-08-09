import { describe, expect, it } from "vitest";

import {
  buildNodeExecutionStatusMap,
  nodeExecutionStatus,
  nodeExecutionStatusLabel,
  parseRunResultNodeIds,
  type Run,
} from "@/lib/node-execution-status";

function run(partial: Partial<Run> & Pick<Run, "id" | "graphId" | "status">): Run {
  return {
    startedAt: "2026-08-09T10:00:00Z",
    ...partial,
  };
}

describe("parseRunResultNodeIds", () => {
  it("extracts nodeIds from result.nodes", () => {
    expect(
      [...parseRunResultNodeIds({
        nodes: [
          { nodeId: "a", value: {} },
          { nodeId: "b", value: {} },
          { value: {} },
        ],
      })].sort(),
    ).toEqual(["a", "b"]);
    expect(parseRunResultNodeIds(null).size).toBe(0);
  });
});

describe("nodeExecutionStatus", () => {
  it("marks waiting from currentWait and completed from result", () => {
    expect(nodeExecutionStatus("n1", null)).toBe("idle");
    expect(
      nodeExecutionStatus(
        "wait-1",
        run({
          id: "r1",
          graphId: "g1",
          status: "running",
          currentWait: { nodeId: "wait-1", signal: "go" },
        }),
      ),
    ).toBe("waiting");
    expect(
      nodeExecutionStatus(
        "start-1",
        run({
          id: "r1",
          graphId: "g1",
          status: "completed",
          result: { nodes: [{ nodeId: "start-1", value: {} }] },
        }),
      ),
    ).toBe("completed");
    expect(
      nodeExecutionStatus(
        "other",
        run({
          id: "r1",
          graphId: "g1",
          status: "completed",
          result: { nodes: [{ nodeId: "start-1", value: {} }] },
        }),
      ),
    ).toBe("idle");
  });

  it("prefers waiting over completed if both somehow present", () => {
    expect(
      nodeExecutionStatus(
        "wait-1",
        run({
          id: "r1",
          graphId: "g1",
          status: "running",
          currentWait: { nodeId: "wait-1", signal: "go" },
          result: { nodes: [{ nodeId: "wait-1", value: {} }] },
        }),
      ),
    ).toBe("waiting");
  });
});

describe("labels and map", () => {
  it("labels statuses for a11y chrome", () => {
    expect(nodeExecutionStatusLabel("waiting")).toBe("Waiting");
    expect(nodeExecutionStatusLabel("completed")).toBe("Done");
    expect(nodeExecutionStatusLabel("idle")).toBeNull();
  });

  it("builds a map for canvas nodes", () => {
    const map = buildNodeExecutionStatusMap(
      ["a", "b"],
      run({
        id: "r1",
        graphId: "g1",
        status: "completed",
        result: { nodes: [{ nodeId: "a", value: 1 }] },
      }),
    );
    expect(map).toEqual({ a: "completed", b: "idle" });
  });
});
