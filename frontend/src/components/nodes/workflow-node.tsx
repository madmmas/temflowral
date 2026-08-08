"use client";

import { Handle, Position, type NodeProps } from "@xyflow/react";

import type { CanvasNode } from "@/lib/graph-canvas";
import { useNodeTypeRegistry } from "@/lib/node-type-registry";
import { resolveOutputHandles } from "@/lib/output-handles";

/**
 * Generic custom node renderer for every palette-created node. Rendering is
 * driven by node data + registry metadata (type/label/category/outputHandles)
 * rather than a per-type switch, so new backend node types render without a
 * matching frontend change (#16, #92).
 */
export function WorkflowNode({ data, selected }: NodeProps<CanvasNode>) {
  const registry = useNodeTypeRegistry();
  const nodeType = registry.find((entry) => entry.id === data.nodeType);
  const outputHandles = resolveOutputHandles(nodeType, data.config);
  const named = outputHandles.length > 0;

  return (
    <div
      className={`relative min-w-36 rounded-md border bg-white px-3 py-2 pb-4 shadow-sm dark:bg-neutral-900 ${
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
