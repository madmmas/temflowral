"use client";

import { useCallback, useEffect, useRef, useState, type PointerEvent } from "react";

import {
  clampRunResultPanelHeight,
  formatRunPayload,
  RUN_RESULT_PANEL_COLLAPSED_HEIGHT,
} from "@/lib/run-result-panel";

type RunResultPanelProps = {
  result?: unknown;
  error?: string | null;
  collapsed: boolean;
  height: number;
  onCollapsedChange: (collapsed: boolean) => void;
  onHeightChange: (height: number) => void;
  onDismiss: () => void;
};

/**
 * Bottom drawer for run.result / run.error (#102). Collapsible, height-
 * resizable, dismissible, with Copy. Keeps JSON out of the footer strip.
 */
export function RunResultPanel({
  result,
  error,
  collapsed,
  height,
  onCollapsedChange,
  onHeightChange,
  onDismiss,
}: RunResultPanelProps) {
  const [copyState, setCopyState] = useState<"idle" | "copied" | "failed">(
    "idle",
  );
  const dragStartY = useRef<number | null>(null);
  const dragStartHeight = useRef(height);

  const bodyText = error
    ? error
    : result !== undefined && result !== null
      ? formatRunPayload(result)
      : "";

  useEffect(() => {
    if (copyState === "idle") return;
    const timer = window.setTimeout(() => setCopyState("idle"), 1_500);
    return () => window.clearTimeout(timer);
  }, [copyState]);

  const onCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(bodyText);
      setCopyState("copied");
    } catch {
      setCopyState("failed");
    }
  }, [bodyText]);

  const onResizePointerDown = useCallback(
    (event: PointerEvent<HTMLDivElement>) => {
      if (collapsed) return;
      event.preventDefault();
      dragStartY.current = event.clientY;
      dragStartHeight.current = height;
      event.currentTarget.setPointerCapture(event.pointerId);
    },
    [collapsed, height],
  );

  const onResizePointerMove = useCallback(
    (event: PointerEvent<HTMLDivElement>) => {
      if (dragStartY.current === null) return;
      // Dragging the top handle upward grows the panel.
      const delta = dragStartY.current - event.clientY;
      onHeightChange(clampRunResultPanelHeight(dragStartHeight.current + delta));
    },
    [onHeightChange],
  );

  const onResizePointerUp = useCallback(
    (event: PointerEvent<HTMLDivElement>) => {
      if (dragStartY.current === null) return;
      dragStartY.current = null;
      if (event.currentTarget.hasPointerCapture(event.pointerId)) {
        event.currentTarget.releasePointerCapture(event.pointerId);
      }
    },
    [],
  );

  const panelHeight = collapsed
    ? RUN_RESULT_PANEL_COLLAPSED_HEIGHT
    : clampRunResultPanelHeight(height);

  return (
    <section
      data-testid="run-result-panel"
      aria-label="Run result"
      className="flex shrink-0 flex-col border-t border-black/10 bg-black/[0.02] dark:border-white/15 dark:bg-white/[0.03]"
      style={{ height: panelHeight }}
    >
      {!collapsed && (
        <div
          role="separator"
          aria-orientation="horizontal"
          aria-label="Resize run result panel"
          data-testid="run-result-resize"
          onPointerDown={onResizePointerDown}
          onPointerMove={onResizePointerMove}
          onPointerUp={onResizePointerUp}
          onPointerCancel={onResizePointerUp}
          className="flex h-2 shrink-0 cursor-ns-resize items-center justify-center touch-none"
        >
          <span className="h-0.5 w-8 rounded-full bg-black/20 dark:bg-white/25" />
        </div>
      )}

      <div className="flex h-9 shrink-0 items-center gap-2 border-b border-black/5 px-3 dark:border-white/10">
        <h2 className="min-w-0 flex-1 truncate text-xs font-semibold uppercase tracking-wide text-black/50 dark:text-white/50">
          {error ? "Run error" : "Run result"}
        </h2>
        <button
          type="button"
          data-testid="run-result-copy"
          onClick={() => void onCopy()}
          disabled={!bodyText}
          className="rounded px-2 py-1 text-xs font-medium text-black/70 hover:bg-black/5 disabled:cursor-not-allowed disabled:opacity-50 dark:text-white/70 dark:hover:bg-white/10"
        >
          {copyState === "copied"
            ? "Copied"
            : copyState === "failed"
              ? "Copy failed"
              : "Copy"}
        </button>
        <button
          type="button"
          data-testid="run-result-collapse"
          aria-expanded={!collapsed}
          onClick={() => onCollapsedChange(!collapsed)}
          className="rounded px-2 py-1 text-xs font-medium text-black/70 hover:bg-black/5 dark:text-white/70 dark:hover:bg-white/10"
        >
          {collapsed ? "Expand" : "Collapse"}
        </button>
        <button
          type="button"
          data-testid="run-result-dismiss"
          onClick={onDismiss}
          className="rounded px-2 py-1 text-xs font-medium text-black/70 hover:bg-black/5 dark:text-white/70 dark:hover:bg-white/10"
        >
          Hide
        </button>
      </div>

      {!collapsed && (
        <div className="min-h-0 flex-1 overflow-auto px-3 py-2">
          {error ? (
            <pre
              data-testid="run-result"
              className="whitespace-pre-wrap break-words font-mono text-xs text-red-600 dark:text-red-400"
            >
              {error}
            </pre>
          ) : (
            <pre
              data-testid="run-result"
              className="whitespace-pre font-mono text-xs text-black/70 dark:text-white/70"
            >
              {bodyText}
            </pre>
          )}
        </div>
      )}
    </section>
  );
}
