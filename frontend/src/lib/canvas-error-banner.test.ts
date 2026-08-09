import { describe, expect, it } from "vitest";

import {
  isCanvasBannerDismissed,
  pickCanvasBannerMessage,
} from "@/lib/canvas-error-banner";

describe("pickCanvasBannerMessage", () => {
  it("prefers action errors over graph-list errors", () => {
    expect(pickCanvasBannerMessage("graph not found", "list failed")).toEqual({
      message: "graph not found",
      source: "action",
    });
    expect(pickCanvasBannerMessage(null, "list failed")).toEqual({
      message: "list failed",
      source: "graphs",
    });
    expect(pickCanvasBannerMessage("  ", null)).toBeNull();
  });
});

describe("isCanvasBannerDismissed", () => {
  it("only hides the same source+message pair", () => {
    const current = { source: "action" as const, message: "graph not found" };
    expect(
      isCanvasBannerDismissed(
        { source: "action", message: "graph not found" },
        current,
      ),
    ).toBe(true);
    expect(
      isCanvasBannerDismissed(
        { source: "action", message: "other" },
        current,
      ),
    ).toBe(false);
    expect(isCanvasBannerDismissed(null, current)).toBe(false);
  });
});
