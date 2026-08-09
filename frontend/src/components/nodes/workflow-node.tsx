"use client";

import { Handle, Position, type NodeProps } from "@xyflow/react";

import type { CanvasNode } from "@/lib/graph-canvas";
import {
  nodeExecutionStatusLabel,
  type NodeExecutionStatus,
} from "@/lib/node-execution-status";
import { useNodeExecutionStatus } from "@/lib/node-execution-status-context";
import { useNodeTypeRegistry } from "@/lib/node-type-registry";
import { resolveOutputHandles } from "@/lib/output-handles";

const statusChrome: Record<
  Exclude<NodeExecutionStatus, "idle">,
  { border: string; badge: string; pulse?: boolean }
> = {
  waiting: {
    border: "border-amber-500 ring-2 ring-amber-500/25",
    badge:
      "bg-amber-500/15 text-amber-900 dark:bg-amber-400/20 dark:text-amber-100",
    pulse: true,
  },
  completed: {
    border: "border-green-500 ring-1 ring-green-500/20",
    badge:
      "bg-green-500/15 text-green-800 dark:bg-green-400/20 dark:text-green-100",
  },
  failed: {
    border: "border-red-500 ring-1 ring-red-500/25",
    badge: "bg-red-500/15 text-red-800 dark:bg-red-400/20 dark:text-red-100",
  },
};

/**
 * Generic custom node renderer for every palette-created node. Rendering is
 * driven by node data + registry metadata (type/label/category/outputHandles)
 * rather than a per-type switch, so new backend node types render without a
 * matching frontend change (#16, #92). Execution chrome comes from the active
 * run (#113).
 */
export function WorkflowNode({ id, data, selected }: NodeProps<CanvasNode>) {
  const registry = useNodeTypeRegistry();
  const nodeType = registry.find((entry) => entry.id === data.nodeType);
  const outputHandles = resolveOutputHandles(nodeType, data.config);
  const named = outputHandles.length > 0;
  const executionStatus = useNodeExecutionStatus(id);
  const statusLabel = nodeExecutionStatusLabel(executionStatus);
  const chrome =
    executionStatus === "idle" ? null : statusChrome[executionStatus];

  const borderClass = selected
    ? "border-blue-500 ring-2 ring-blue-500/30"
    : chrome
      ? chrome.border
      : "border-black/15 dark:border-white/20";

  return (
    <div
      data-testid={`workflow-node-${id}`}
      data-execution-status={executionStatus}
      className={`relative min-w-36 rounded-md border bg-white px-3 py-2 pb-4 shadow-sm dark:bg-neutral-900 ${borderClass} ${
        chrome?.pulse ? "animate-pulse" : ""
      }`}
    >
      <Handle type="target" position={Position.Top} />
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <div className="text-sm font-medium text-black dark:text-white">
            {data.label}
          </div>
          <div className="mt-0.5 text-[10px] uppercase tracking-wide text-black/40 dark:text-white/40">
            {data.category
              ? `${data.category} · ${data.nodeType}`
              : data.nodeType}
          </div>
        </div>
        {statusLabel && chrome && (
          <span
            data-testid={`node-execution-status-${id}`}
            className={`shrink-0 rounded px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${chrome.badge}`}
          >
            {statusLabel}
          </span>
        )}
      </div>
      {named ? (
        <>
          <div className="mt-2 flex justify-between gap-1 px-0.5">
            {outputHandles.map((handle) => (
              <span
                key={`label-${handle.id}`}
                className="min-w-0 flex-1 truncate text-center text-[9px] text-black/45 dark:text-white/45"
              >
                {handle.label ?? handle.id}
              </span>
            ))}
          </div>
          {outputHandles.map((handle, index) => {
            const leftPercent =
              ((index + 1) / (outputHandles.length + 1)) * 100;
            return (
              <Handle
                key={handle.id}
                type="source"
                position={Position.Bottom}
                id={handle.id}
                style={{ left: `${leftPercent}%` }}
                data-testid={`source-handle-${handle.id}`}
              />
            );
          })}
        </>
      ) : (
        <Handle type="source" position={Position.Bottom} />
      )}
    </div>
  );
}
