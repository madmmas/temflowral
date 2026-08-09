import type { components } from "@/api";

export type Run = components["schemas"]["Run"];
export type RunStatus = components["schemas"]["RunStatus"];

/** Session-local snapshot of a run for the canvas history list (#106). */
export type RunHistoryEntry = {
  id: string;
  graphId: string;
  status: RunStatus;
  startedAt: string;
  completedAt?: string;
  result?: unknown;
  error?: string | null;
  currentWait?: Run["currentWait"];
};

export const RUN_HISTORY_MAX = 20;

export function runToHistoryEntry(run: Run): RunHistoryEntry {
  return {
    id: run.id,
    graphId: run.graphId,
    status: run.status,
    startedAt: run.startedAt,
    completedAt: run.completedAt,
    result: run.result,
    error: run.error,
    currentWait: run.currentWait,
  };
}

/** Rebuild a Run-shaped object from a session history entry (no live Temporal). */
export function historyEntryToRun(entry: RunHistoryEntry): Run {
  return {
    id: entry.id,
    graphId: entry.graphId,
    status: entry.status,
    startedAt: entry.startedAt,
    completedAt: entry.completedAt,
    result: entry.result as Run["result"],
    error: entry.error ?? undefined,
    currentWait: entry.currentWait,
  };
}

/**
 * Insert or replace a run by id at the front of the list (newest first).
 * Caps length so the session footer stays usable.
 */
export function upsertRunHistory(
  history: readonly RunHistoryEntry[],
  entry: RunHistoryEntry,
  max: number = RUN_HISTORY_MAX,
): RunHistoryEntry[] {
  const next = [entry, ...history.filter((item) => item.id !== entry.id)];
  return next.slice(0, Math.max(1, max));
}

export function runsForGraph(
  history: readonly RunHistoryEntry[],
  graphId: string | null,
): RunHistoryEntry[] {
  if (!graphId) return [];
  return history.filter((entry) => entry.graphId === graphId);
}

export function formatRunStatusLabel(status: RunStatus): string {
  return `Run ${status}`;
}

export function runStatusTone(
  status: RunStatus,
): "success" | "danger" | "info" {
  if (status === "completed") return "success";
  if (status === "failed" || status === "cancelled") return "danger";
  return "info";
}

/** Compact local time for history rows. */
export function formatRunTimestamp(iso: string): string {
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  return new Intl.DateTimeFormat(undefined, {
    month: "short",
    day: "numeric",
    hour: "numeric",
    minute: "2-digit",
  }).format(date);
}

export function runHasInspectablePayload(entry: {
  result?: unknown;
  error?: string | null;
}): boolean {
  if (entry.error != null && entry.error !== "") return true;
  return entry.result !== undefined && entry.result !== null;
}
