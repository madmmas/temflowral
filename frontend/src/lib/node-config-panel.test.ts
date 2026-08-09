import { describe, expect, it } from "vitest";

import {
  clampNodeConfigPanelWidth,
  NODE_CONFIG_PANEL_DEFAULT_WIDTH,
  NODE_CONFIG_PANEL_MAX_WIDTH,
  NODE_CONFIG_PANEL_MIN_WIDTH,
  NODE_CONFIG_WIDTH_STORAGE_KEY,
  readNodeConfigPanelWidth,
  writeNodeConfigPanelWidth,
} from "@/lib/node-config-panel";

function memoryStorage(initial: Record<string, string> = {}) {
  const store = { ...initial };
  return {
    getItem: (key: string) => (key in store ? store[key] : null),
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    _store: store,
  };
}

describe("clampNodeConfigPanelWidth", () => {
  it("clamps to min/max and rounds", () => {
    expect(clampNodeConfigPanelWidth(10)).toBe(NODE_CONFIG_PANEL_MIN_WIDTH);
    expect(clampNodeConfigPanelWidth(9999)).toBe(NODE_CONFIG_PANEL_MAX_WIDTH);
    expect(clampNodeConfigPanelWidth(300.6)).toBe(301);
    expect(clampNodeConfigPanelWidth(Number.NaN)).toBe(
      NODE_CONFIG_PANEL_DEFAULT_WIDTH,
    );
  });
});

describe("node config width storage", () => {
  it("reads and writes a clamped width", () => {
    const storage = memoryStorage();
    writeNodeConfigPanelWidth(320, storage);
    expect(storage._store[NODE_CONFIG_WIDTH_STORAGE_KEY]).toBe("320");
    expect(readNodeConfigPanelWidth(storage)).toBe(320);
  });

  it("falls back when missing or invalid", () => {
    expect(readNodeConfigPanelWidth(null)).toBe(NODE_CONFIG_PANEL_DEFAULT_WIDTH);
    expect(readNodeConfigPanelWidth(memoryStorage())).toBe(
      NODE_CONFIG_PANEL_DEFAULT_WIDTH,
    );
    expect(
      readNodeConfigPanelWidth(
        memoryStorage({ [NODE_CONFIG_WIDTH_STORAGE_KEY]: "nope" }),
      ),
    ).toBe(NODE_CONFIG_PANEL_DEFAULT_WIDTH);
  });
});
