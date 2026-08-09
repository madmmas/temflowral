import { describe, expect, it } from "vitest";

import {
  AUTHORING_TIP_DISMISSED_STORAGE_KEY,
  paletteClickScreenOffset,
  readAuthoringTipDismissed,
  writeAuthoringTipDismissed,
} from "@/lib/canvas-guidance";

function memoryStorage(initial: Record<string, string> = {}) {
  const store = { ...initial };
  return {
    getItem: (key: string) => (key in store ? store[key] : null),
    setItem: (key: string, value: string) => {
      store[key] = value;
    },
    removeItem: (key: string) => {
      delete store[key];
    },
    _store: store,
  };
}

describe("authoring tip dismissal", () => {
  it("defaults to not dismissed", () => {
    expect(readAuthoringTipDismissed(null)).toBe(false);
    expect(readAuthoringTipDismissed(memoryStorage())).toBe(false);
  });

  it("persists dismissal and can clear it", () => {
    const storage = memoryStorage();
    writeAuthoringTipDismissed(true, storage);
    expect(storage._store[AUTHORING_TIP_DISMISSED_STORAGE_KEY]).toBe("true");
    expect(readAuthoringTipDismissed(storage)).toBe(true);

    writeAuthoringTipDismissed(false, storage);
    expect(storage._store[AUTHORING_TIP_DISMISSED_STORAGE_KEY]).toBeUndefined();
    expect(readAuthoringTipDismissed(storage)).toBe(false);
  });
});

describe("paletteClickScreenOffset", () => {
  it("fans out successive center drops", () => {
    expect(paletteClickScreenOffset(0)).toEqual({ x: 0, y: 0 });
    expect(paletteClickScreenOffset(1)).toEqual({ x: 32, y: 0 });
    expect(paletteClickScreenOffset(4)).toEqual({ x: 0, y: 32 });
  });
});
