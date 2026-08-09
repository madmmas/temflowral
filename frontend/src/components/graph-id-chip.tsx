"use client";

import { useEffect, useState } from "react";

import { graphShortId } from "@/lib/workflow-library";

type GraphIdChipProps = {
  graphId: string;
};

/** Short graph id with copy — avoids dumping the full UUID in the footer (#106). */
export function GraphIdChip({ graphId }: GraphIdChipProps) {
  const [copyState, setCopyState] = useState<"idle" | "copied" | "failed">(
    "idle",
  );

  useEffect(() => {
    if (copyState === "idle") return;
    const timer = window.setTimeout(() => setCopyState("idle"), 1_500);
    return () => window.clearTimeout(timer);
  }, [copyState]);

  const onCopy = async () => {
    try {
      await navigator.clipboard.writeText(graphId);
      setCopyState("copied");
    } catch {
      setCopyState("failed");
    }
  };

  return (
    <span
      data-testid="graph-id-chip"
      className="inline-flex items-center gap-1.5 text-black/50 dark:text-white/50"
      title={graphId}
    >
      <span>
        Graph: <span className="font-mono">{graphShortId(graphId)}</span>
      </span>
      <button
        type="button"
        data-testid="copy-graph-id"
        onClick={() => void onCopy()}
        className="rounded px-1.5 py-0.5 text-[10px] font-medium text-black/60 hover:bg-black/5 dark:text-white/60 dark:hover:bg-white/10"
      >
        {copyState === "copied"
          ? "Copied"
          : copyState === "failed"
            ? "Copy failed"
            : "Copy"}
      </button>
    </span>
  );
}
