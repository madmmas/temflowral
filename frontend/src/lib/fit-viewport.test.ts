import { describe, expect, it } from "vitest";

import { shouldFitViewportAfterLoad } from "@/lib/fit-viewport";

describe("shouldFitViewportAfterLoad", () => {
  it("is true only when there is at least one node", () => {
    expect(shouldFitViewportAfterLoad(0)).toBe(false);
    expect(shouldFitViewportAfterLoad(1)).toBe(true);
    expect(shouldFitViewportAfterLoad(3)).toBe(true);
    expect(shouldFitViewportAfterLoad(Number.NaN)).toBe(false);
  });
});
