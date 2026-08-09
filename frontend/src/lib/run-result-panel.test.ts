import { describe, expect, it } from "vitest";

import {
  clampRunResultPanelHeight,
  formatRunPayload,
  runHasResultPanelContent,
  RUN_RESULT_PANEL_DEFAULT_HEIGHT,
  RUN_RESULT_PANEL_MAX_HEIGHT,
  RUN_RESULT_PANEL_MIN_HEIGHT,
} from "@/lib/run-result-panel";

describe("formatRunPayload", () => {
  it("pretty-prints objects as JSON", () => {
    expect(formatRunPayload({ message: "ok" })).toBe(
      '{\n  "message": "ok"\n}',
    );
  });

  it("returns strings as-is", () => {
    expect(formatRunPayload("plain")).toBe("plain");
  });
});

describe("clampRunResultPanelHeight", () => {
  it("clamps below min and above max", () => {
    expect(clampRunResultPanelHeight(10)).toBe(RUN_RESULT_PANEL_MIN_HEIGHT);
    expect(clampRunResultPanelHeight(9999)).toBe(RUN_RESULT_PANEL_MAX_HEIGHT);
  });

  it("rounds finite values in range", () => {
    expect(clampRunResultPanelHeight(150.7)).toBe(151);
  });

  it("falls back for non-finite input", () => {
    expect(clampRunResultPanelHeight(Number.NaN)).toBe(
      RUN_RESULT_PANEL_DEFAULT_HEIGHT,
    );
  });
});

describe("runHasResultPanelContent", () => {
  it("is false without a run", () => {
    expect(runHasResultPanelContent(null)).toBe(false);
  });

  it("is true for result or error", () => {
    expect(runHasResultPanelContent({ result: { ok: true } })).toBe(true);
    expect(runHasResultPanelContent({ error: "boom" })).toBe(true);
  });

  it("is false when both are empty", () => {
    expect(runHasResultPanelContent({})).toBe(false);
    expect(runHasResultPanelContent({ error: "" })).toBe(false);
  });
});
