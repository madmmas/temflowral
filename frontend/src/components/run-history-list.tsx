"use client";

import {
  formatRunTimestamp,
  runHasInspectablePayload,
  type RunHistoryEntry,
} from "@/lib/run-history";
import { graphShortId } from "@/lib/workflow-library";

type RunHistoryListProps = {
  entries: RunHistoryEntry[];
  activeRunId: string | null;
  onSelect: (entry: RunHistoryEntry) => void;
};

/**
 * Session-local recent runs for the open graph (#106). Clicking an entry
 * restores that run's result/error in the drawer.
 */
export function RunHistoryList({
  entries,
  activeRunId,
  onSelect,
}: RunHistoryListProps) {
  if (entries.length === 0) return null;

  return (
    <div
      data-testid="run-history"
      className="flex min-w-0 flex-1 flex-wrap items-center gap-1.5"
    >
      <span className="shrink-0 text-[10px] font-semibold uppercase tracking-wide text-black/40 dark:text-white/40">
        Recent
      </span>
      <ul className="flex min-w-0 flex-wrap items-center gap-1">
        {entries.map((entry) => {
          const active = entry.id === activeRunId;
          const stamp = formatRunTimestamp(
            entry.completedAt ?? entry.startedAt,
          );
          const inspectable = runHasInspectablePayload(entry);
          return (
            <li key={entry.id}>
              <button
                type="button"
                data-testid={`run-history-item-${entry.id}`}
                disabled={!inspectable && entry.status === "running"}
                onClick={() => onSelect(entry)}
                title={`${entry.status} · ${entry.id}`}
                className={`rounded-md border px-2 py-0.5 font-mono text-[10px] disabled:cursor-not-allowed disabled:opacity-50 ${
                  active
                    ? "border-blue-500/50 bg-blue-500/10 text-blue-800 dark:text-blue-200"
                    : "border-black/10 bg-white text-black/65 hover:bg-black/5 dark:border-white/15 dark:bg-neutral-900 dark:text-white/65 dark:hover:bg-white/10"
                }`}
              >
                {graphShortId(entry.id)} · {entry.status}
                <span className="ml-1 text-black/40 dark:text-white/40">
                  {stamp}
                </span>
              </button>
            </li>
          );
        })}
      </ul>
    </div>
  );
}
