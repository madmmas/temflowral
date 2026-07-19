"use client";

import { Handle, Position, type NodeProps } from "@xyflow/react";

import type { CanvasNode } from "@/lib/graph-canvas";

/**
 * Generic custom node renderer for every palette-created node. Rendering is
 * driven by node data (type/label/category) rather than a per-type switch, so
 * new backend node types render without a matching frontend change (#16).
 * Type-specific rendering can be added as sibling components later.
 */
export function WorkflowNode({ data, selected }: NodeProps<CanvasNode>) {
  return (
    <div
      className={`min-w-32 rounded-md border bg-white px-3 py-2 shadow-sm dark:bg-neutral-900 ${
        selected
          ? "border-blue-500 ring-2 ring-blue-500/30"
          : "border-black/15 dark:border-white/20"
      }`}
    >
      <Handle type="target" position={Position.Top} />
      <div className="text-sm font-medium text-black dark:text-white">
        {data.label}
      </div>
      <div className="mt-0.5 text-[10px] uppercase tracking-wide text-black/40 dark:text-white/40">
        {data.category ? `${data.category} · ${data.nodeType}` : data.nodeType}
      </div>
      <Handle type="source" position={Position.Bottom} />
    </div>
  );
}
