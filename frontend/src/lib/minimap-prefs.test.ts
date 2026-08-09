import { describe, expect, it } from "vitest";

import {
  MINIMAP_VISIBLE_DEFAULT,
  MINIMAP_VISIBLE_STORAGE_KEY,
  minimapColorsForScheme,
  MINIMAP_COLORS_DARK,
  MINIMAP_COLORS_LIGHT,
  readMinimapVisible,
  writeMinimapVisible,
} from "@/lib/minimap-prefs";

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

describe("readMinimapVisible", () => {
  it("returns the default when storage is empty or unavailable", () => {
    expect(readMinimapVisible(null)).toBe(MINIMAP_VISIBLE_DEFAULT);
    expect(readMinimapVisible(memoryStorage())).toBe(MINIMAP_VISIBLE_DEFAULT);
  });

  it("parses stored true/false", () => {
    expect(
      readMinimapVisible(
        memoryStorage({ [MINIMAP_VISIBLE_STORAGE_KEY]: "false" }),
      ),
    ).toBe(false);
    expect(
      readMinimapVisible(
        memoryStorage({ [MINIMAP_VISIBLE_STORAGE_KEY]: "true" }),
      ),
    ).toBe(true);
  });

  it("falls back on unexpected values", () => {
    expect(
      readMinimapVisible(
        memoryStorage({ [MINIMAP_VISIBLE_STORAGE_KEY]: "maybe" }),
      ),
    ).toBe(MINIMAP_VISIBLE_DEFAULT);
  });
});

describe("writeMinimapVisible", () => {
  it("persists the preference", () => {
    const storage = memoryStorage();
    writeMinimapVisible(false, storage);
    expect(storage._store[MINIMAP_VISIBLE_STORAGE_KEY]).toBe("false");
    writeMinimapVisible(true, storage);
    expect(storage._store[MINIMAP_VISIBLE_STORAGE_KEY]).toBe("true");
  });

  it("no-ops when storage is null", () => {
    expect(() => writeMinimapVisible(true, null)).not.toThrow();
  });
});

describe("minimapColorsForScheme", () => {
  it("returns dark palette for dark preference", () => {
    expect(minimapColorsForScheme(true)).toBe(MINIMAP_COLORS_DARK);
    expect(minimapColorsForScheme(false)).toBe(MINIMAP_COLORS_LIGHT);
  });
});
