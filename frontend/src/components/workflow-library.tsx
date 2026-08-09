"use client";

import { useEffect, useId, useMemo, useRef, useState } from "react";

import type { GraphSummary } from "@/lib/graph-canvas";
import {
  filterAndSortGraphs,
  formatGraphUpdatedAt,
  graphDisplayName,
  graphShortId,
  type WorkflowLibrarySort,
} from "@/lib/workflow-library";

type WorkflowLibraryProps = {
  open: boolean;
  graphs: GraphSummary[];
  loading: boolean;
  error: string | null;
  currentGraphId: string | null;
  busy?: boolean;
  onClose: () => void;
  onOpenGraph: (graphId: string) => void;
  onRefresh: () => void;
};

/**
 * Searchable modal picker for saved graphs (#103). Replaces the toolbar
 * `<select>` so duplicate names stay distinguishable via updated time + id.
 */
export function WorkflowLibrary({
  open,
  graphs,
  loading,
  error,
  currentGraphId,
  busy = false,
  onClose,
  onOpenGraph,
  onRefresh,
}: WorkflowLibraryProps) {
  const titleId = useId();
  const searchRef = useRef<HTMLInputElement>(null);
  const [query, setQuery] = useState("");
  const [sort, setSort] = useState<WorkflowLibrarySort>("updatedAt");

  useEffect(() => {
    if (!open) return;
    setQuery("");
    setSort("updatedAt");
    onRefresh();
    const timer = window.setTimeout(() => searchRef.current?.focus(), 0);
    return () => window.clearTimeout(timer);
  }, [open, onRefresh]);

  useEffect(() => {
    if (!open) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        event.preventDefault();
        onClose();
      }
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [open, onClose]);

  const visible = useMemo(
    () => filterAndSortGraphs(graphs, query, sort),
    [graphs, query, sort],
  );

  if (!open) return null;

  return (
    <div
      data-testid="workflow-library"
      className="fixed inset-0 z-50 flex items-start justify-center bg-black/40 px-4 py-16 dark:bg-black/60"
      onMouseDown={(event) => {
        if (event.target === event.currentTarget) onClose();
      }}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby={titleId}
        className="flex max-h-[min(32rem,calc(100vh-8rem))] w-full max-w-lg flex-col overflow-hidden rounded-lg border border-black/10 bg-white shadow-lg dark:border-white/15 dark:bg-neutral-950"
      >
        <div className="flex items-start justify-between gap-3 border-b border-black/10 px-4 py-3 dark:border-white/15">
          <div>
            <h2
              id={titleId}
              className="text-sm font-semibold text-black dark:text-white"
            >
              Open workflow
            </h2>
            <p className="mt-0.5 text-[11px] text-black/50 dark:text-white/50">
              Search by name or id · sorted list
            </p>
          </div>
          <button
            type="button"
            data-testid="workflow-library-close"
            onClick={onClose}
            className="rounded px-2 py-1 text-xs text-black/60 hover:bg-black/5 dark:text-white/60 dark:hover:bg-white/10"
          >
            Close
          </button>
        </div>

        <div className="flex flex-wrap items-center gap-2 border-b border-black/5 px-4 py-2 dark:border-white/10">
          <input
            ref={searchRef}
            data-testid="workflow-library-search"
            aria-label="Search workflows"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search name or id…"
            className="min-w-40 flex-1 rounded-md border border-black/10 bg-white px-2.5 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
          />
          <label className="flex items-center gap-1.5 text-[11px] text-black/55 dark:text-white/55">
            Sort
            <select
              data-testid="workflow-library-sort"
              aria-label="Sort workflows"
              value={sort}
              onChange={(event) =>
                setSort(event.target.value as WorkflowLibrarySort)
              }
              className="rounded-md border border-black/10 bg-white px-2 py-1.5 text-xs dark:border-white/15 dark:bg-neutral-900"
            >
              <option value="updatedAt">Updated</option>
              <option value="name">Name</option>
            </select>
          </label>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto">
          {loading && (
            <p className="px-4 py-6 text-center text-xs text-black/50 dark:text-white/50">
              Loading workflows…
            </p>
          )}
          {!loading && error && (
            <p className="px-4 py-6 text-center text-xs text-red-600 dark:text-red-400">
              {error}
            </p>
          )}
          {!loading && !error && visible.length === 0 && (
            <p className="px-4 py-6 text-center text-xs text-black/50 dark:text-white/50">
              {graphs.length === 0
                ? "No saved workflows yet."
                : "No workflows match your search."}
            </p>
          )}
          {!loading && !error && visible.length > 0 && (
            <ul className="divide-y divide-black/5 dark:divide-white/10">
              {visible.map((graph) => {
                const current = graph.id === currentGraphId;
                return (
                  <li key={graph.id}>
                    <button
                      type="button"
                      data-testid={`workflow-library-item-${graph.id}`}
                      disabled={busy}
                      onClick={() => onOpenGraph(graph.id)}
                      className={`flex w-full flex-col gap-0.5 px-4 py-2.5 text-left hover:bg-black/[0.04] disabled:cursor-not-allowed disabled:opacity-50 dark:hover:bg-white/[0.06] ${
                        current ? "bg-blue-500/10" : ""
                      }`}
                    >
                      <span className="flex items-center gap-2 text-sm font-medium text-black dark:text-white">
                        <span className="min-w-0 flex-1 truncate">
                          {graphDisplayName(graph)}
                        </span>
                        {current && (
                          <span className="shrink-0 text-[10px] font-semibold uppercase tracking-wide text-blue-700 dark:text-blue-300">
                            Current
                          </span>
                        )}
                      </span>
                      <span className="flex flex-wrap items-center gap-x-2 gap-y-0.5 font-mono text-[11px] text-black/45 dark:text-white/45">
                        <span>{graphShortId(graph.id)}</span>
                        <span aria-hidden="true">·</span>
                        <span>Updated {formatGraphUpdatedAt(graph.updatedAt)}</span>
                      </span>
                    </button>
                  </li>
                );
              })}
            </ul>
          )}
        </div>
      </div>
    </div>
  );
}
