import { describe, expect, it } from "vitest";

import { resolveOutputHandles, type NodeType } from "./output-handles";

const conditionType: NodeType = {
  id: "condition",
  name: "Condition",
  configSchema: {},
  outputHandles: [
    { id: "true", label: "True" },
    { id: "false", label: "False" },
  ],
};

const waitType: NodeType = {
  id: "wait",
  name: "Wait for Signal",
  configSchema: {},
  outputHandles: [
    { id: "received", label: "Received" },
    { id: "timedOut", label: "Timed out" },
  ],
};

const branchType: NodeType = {
  id: "switch",
  name: "Switch",
  configSchema: {},
  outputHandlesFromConfig: { path: "branches" },
};

describe("resolveOutputHandles", () => {
  it("returns fixed handles for condition and wait", () => {
    expect(resolveOutputHandles(conditionType, {})).toEqual([
      { id: "true", label: "True" },
      { id: "false", label: "False" },
    ]);
    expect(resolveOutputHandles(waitType, {})).toEqual([
      { id: "received", label: "Received" },
      { id: "timedOut", label: "Timed out" },
    ]);
  });

  it("returns an empty list for types without named outputs", () => {
    expect(
      resolveOutputHandles(
        { id: "noop", name: "No-op", configSchema: {} },
        {},
      ),
    ).toEqual([]);
    expect(resolveOutputHandles(undefined, {})).toEqual([]);
  });

  it("derives handles from config arrays and object keys", () => {
    expect(
      resolveOutputHandles(branchType, {
        branches: [{ id: "a" }, { id: "b" }],
      }),
    ).toEqual([{ id: "a" }, { id: "b" }]);
    expect(
      resolveOutputHandles(branchType, {
        branches: ["left", "right"],
      }),
    ).toEqual([{ id: "left" }, { id: "right" }]);
    expect(
      resolveOutputHandles(
        { ...branchType, outputHandlesFromConfig: { path: "routes" } },
        { routes: { z: 1, a: 2 } },
      ),
    ).toEqual([{ id: "a" }, { id: "z" }]);
  });
});
