"use client";

import { groupByCategory, useNodeTypes, type NodeType } from "@/lib/node-types";

/** MIME-ish key used to carry a node-type id through an HTML5 drag. */
export const NODE_TYPE_DRAG_KEY = "application/temflowral-node-type";

type NodePaletteProps = {
  /** Add a node of this type at the viewport centre (click fallback for DnD). */
  onAddNodeType: (nodeType: NodeType) => void;
};

/**
 * Sidebar palette populated from `GET /node-types`. Drag an item onto the
 * canvas, or click it to drop one at the centre.
 */
export function NodePalette({ onAddNodeType }: NodePaletteProps) {
  const { nodeTypes, loading, error } = useNodeTypes();

  return (
    <aside
      data-testid="node-palette"
      className="flex w-56 shrink-0 flex-col gap-3 overflow-y-auto border-r border-black/10 bg-black/[0.02] p-3 dark:border-white/15 dark:bg-white/[0.03]"
    >
      <h2 className="text-xs font-semibold uppercase tracking-wide text-black/50 dark:text-white/50">
        Nodes
      </h2>

      {loading && (
        <p className="text-xs text-black/50 dark:text-white/50">Loading…</p>
      )}

      {error && (
        <p className="text-xs text-red-600 dark:text-red-400">{error}</p>
      )}

      {!loading && !error && nodeTypes.length === 0 && (
        <p className="text-xs text-black/50 dark:text-white/50">
          No node types available.
        </p>
      )}

      {groupByCategory(nodeTypes).map(([category, types]) => (
        <div key={category} className="flex flex-col gap-1.5">
          <h3 className="text-[10px] font-semibold uppercase tracking-wide text-black/40 dark:text-white/40">
            {category}
          </h3>
          {types.map((nodeType) => (
            <button
              key={nodeType.id}
              data-testid={`node-type-${nodeType.id}`}
              type="button"
              draggable
              onDragStart={(event) => {
                event.dataTransfer.setData(NODE_TYPE_DRAG_KEY, nodeType.id);
                event.dataTransfer.effectAllowed = "move";
              }}
              onClick={() => onAddNodeType(nodeType)}
              title={nodeType.description ?? nodeType.name}
              className="cursor-grab rounded-md border border-black/10 bg-white px-2.5 py-1.5 text-left text-sm shadow-sm hover:bg-black/5 active:cursor-grabbing dark:border-white/15 dark:bg-neutral-900 dark:hover:bg-white/10"
            >
              {nodeType.name}
            </button>
          ))}
        </div>
      ))}
    </aside>
  );
}
