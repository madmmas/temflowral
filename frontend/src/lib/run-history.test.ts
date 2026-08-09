import { describe, expect, it } from "vitest";

import {
  formatRunStatusLabel,
  historyEntryToRun,
  runHasInspectablePayload,
  runsForGraph,
  runStatusTone,
  runToHistoryEntry,
  upsertRunHistory,
  type Run,
  type RunHistoryEntry,
} from "@/lib/run-history";

function run(partial: Partial<Run> & Pick<Run, "id" | "graphId" | "status">): Run {
  return {
    startedAt: "2026-08-09T10:00:00Z",
    ...partial,
  };
}

describe("upsertRunHistory", () => {
  it("prepends new runs and updates existing by id", () => {
    const first = runToHistoryEntry(
      run({
        id: "r1",
        graphId: "g1",
        status: "running",
      }),
    );
    const second = runToHistoryEntry(
      run({
        id: "r2",
        graphId: "g1",
        status: "completed",
        result: { ok: true },
      }),
    );
    let history = upsertRunHistory([], first);
    history = upsertRunHistory(history, second);
    expect(history.map((e) => e.id)).toEqual(["r2", "r1"]);

    history = upsertRunHistory(history, {
      ...first,
      status: "completed",
      result: { done: 1 },
    });
    expect(history[0]?.id).toBe("r1");
    expect(history[0]?.status).toBe("completed");
    expect(history).toHaveLength(2);
  });

  it("caps history length", () => {
    let history: RunHistoryEntry[] = [];
    for (let i = 0; i < 5; i++) {
      history = upsertRunHistory(
        history,
        runToHistoryEntry(
          run({ id: `r${i}`, graphId: "g1", status: "completed" }),
        ),
        3,
      );
    }
    expect(history).toHaveLength(3);
    expect(history.map((e) => e.id)).toEqual(["r4", "r3", "r2"]);
  });
});

describe("runsForGraph", () => {
  it("filters by graph id", () => {
    const history = [
      runToHistoryEntry(run({ id: "a", graphId: "g1", status: "completed" })),
      runToHistoryEntry(run({ id: "b", graphId: "g2", status: "failed" })),
    ];
    expect(runsForGraph(history, "g1").map((e) => e.id)).toEqual(["a"]);
    expect(runsForGraph(history, null)).toEqual([]);
  });
});

describe("status helpers", () => {
  it("labels and tones match status", () => {
    expect(formatRunStatusLabel("completed")).toBe("Run completed");
    expect(runStatusTone("completed")).toBe("success");
    expect(runStatusTone("failed")).toBe("danger");
    expect(runStatusTone("running")).toBe("info");
    expect(runStatusTone("pending")).toBe("info");
  });
});

describe("historyEntryToRun / inspectable", () => {
  it("round-trips enough fields to reopen a result", () => {
    const entry = runToHistoryEntry(
      run({
        id: "r1",
        graphId: "g1",
        status: "completed",
        result: { message: "ok" },
      }),
    );
    expect(historyEntryToRun(entry).result).toEqual({ message: "ok" });
    expect(runHasInspectablePayload(entry)).toBe(true);
    expect(runHasInspectablePayload({ error: "" })).toBe(false);
  });
});
