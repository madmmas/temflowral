import { describe, expect, it, vi } from "vitest";

import type { GraphSummary } from "@/lib/graph-canvas";
import {
  DEFAULT_GRAPH_NAME,
  duplicateNameConfirmMessage,
  findGraphsWithSameName,
  graphNameHint,
  isPlaceholderGraphName,
  normalizeGraphName,
  resolveGraphNameForSave,
} from "@/lib/graph-naming";

function summary(
  partial: Partial<GraphSummary> & Pick<GraphSummary, "id" | "name">,
): GraphSummary {
  return {
    createdAt: "2026-08-09T00:00:00Z",
    updatedAt: "2026-08-09T00:00:00Z",
    ...partial,
  };
}

describe("isPlaceholderGraphName / normalize", () => {
  it("treats blank and default Untitled as placeholders", () => {
    expect(isPlaceholderGraphName("")).toBe(true);
    expect(isPlaceholderGraphName("   ")).toBe(true);
    expect(isPlaceholderGraphName(DEFAULT_GRAPH_NAME)).toBe(true);
    expect(isPlaceholderGraphName("untitled workflow")).toBe(true);
    expect(isPlaceholderGraphName("My flow")).toBe(false);
  });

  it("normalizes blank to the default", () => {
    expect(normalizeGraphName("  ")).toBe(DEFAULT_GRAPH_NAME);
    expect(normalizeGraphName("  Hello  ")).toBe("Hello");
  });
});

describe("findGraphsWithSameName", () => {
  const graphs = [
    summary({ id: "a", name: "Alpha" }),
    summary({ id: "b", name: "alpha" }),
    summary({ id: "c", name: "Beta" }),
  ];

  it("matches case-insensitively and can exclude self", () => {
    expect(findGraphsWithSameName(graphs, "ALPHA").map((g) => g.id)).toEqual([
      "a",
      "b",
    ]);
    expect(
      findGraphsWithSameName(graphs, "Alpha", "a").map((g) => g.id),
    ).toEqual(["b"]);
    expect(findGraphsWithSameName(graphs, "missing")).toEqual([]);
  });
});

describe("resolveGraphNameForSave", () => {
  const graphs = [summary({ id: "a", name: "Existing" })];

  it("prompts on first save when name is placeholder", () => {
    const prompt = vi.fn().mockReturnValue("Checkout flow");
    const confirm = vi.fn().mockReturnValue(true);
    const result = resolveGraphNameForSave({
      name: DEFAULT_GRAPH_NAME,
      isNewGraph: true,
      graphs: [],
      prompt,
      confirm,
    });
    expect(result).toEqual({ ok: true, name: "Checkout flow" });
    expect(prompt).toHaveBeenCalled();
    expect(confirm).not.toHaveBeenCalled();
  });

  it("cancels when the name prompt is dismissed", () => {
    const result = resolveGraphNameForSave({
      name: DEFAULT_GRAPH_NAME,
      isNewGraph: true,
      graphs: [],
      prompt: () => null,
      confirm: () => true,
    });
    expect(result).toEqual({ ok: false, reason: "cancelled" });
  });

  it("rejects prompted placeholder names", () => {
    const result = resolveGraphNameForSave({
      name: DEFAULT_GRAPH_NAME,
      isNewGraph: true,
      graphs: [],
      prompt: () => "  Untitled workflow  ",
      confirm: () => true,
    });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.reason).toBe("placeholder");
  });

  it("does not prompt for already-saved graphs keeping Untitled", () => {
    const prompt = vi.fn();
    const result = resolveGraphNameForSave({
      name: DEFAULT_GRAPH_NAME,
      isNewGraph: false,
      graphs: [],
      excludeGraphId: "x",
      prompt,
      confirm: () => true,
    });
    expect(result).toEqual({ ok: true, name: DEFAULT_GRAPH_NAME });
    expect(prompt).not.toHaveBeenCalled();
  });

  it("soft-confirms duplicate names", () => {
    const confirm = vi.fn().mockReturnValue(false);
    const result = resolveGraphNameForSave({
      name: "Existing",
      isNewGraph: true,
      graphs,
      prompt: () => null,
      confirm,
    });
    expect(result).toEqual({ ok: false, reason: "duplicate_declined" });
    expect(confirm).toHaveBeenCalledWith(
      duplicateNameConfirmMessage("Existing", 1),
    );
  });

  it("accepts duplicates when confirmed", () => {
    const result = resolveGraphNameForSave({
      name: "Existing",
      isNewGraph: false,
      graphs,
      excludeGraphId: "other",
      prompt: () => null,
      confirm: () => true,
    });
    expect(result).toEqual({ ok: true, name: "Existing" });
  });
});

describe("graphNameHint", () => {
  it("nudges placeholder on new graphs and warns on duplicates", () => {
    expect(
      graphNameHint({
        name: DEFAULT_GRAPH_NAME,
        isNewGraph: true,
        duplicateCount: 0,
      }),
    ).toMatch(/Name this workflow/);
    expect(
      graphNameHint({
        name: "Existing",
        isNewGraph: false,
        duplicateCount: 2,
      }),
    ).toMatch(/2 saved workflows/);
    expect(
      graphNameHint({
        name: "Unique",
        isNewGraph: false,
        duplicateCount: 0,
      }),
    ).toBeNull();
  });
});
