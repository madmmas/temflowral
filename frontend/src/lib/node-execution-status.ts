import type { components } from "@/api";

export type Run = components["schemas"]["Run"];

/**
 * Visual execution state for a canvas node (#113).
 *
 * Interim (no mid-run progress API yet):
 * - `completed` from terminal `run.result.nodes[].nodeId`
 * - `waiting` from `run.currentWait.nodeId` while status is running
 * - otherwise `idle` (full live pending/running needs a contract progress field)
 */
export type NodeExecutionStatus = "idle" | "waiting" | "completed" | "failed";

export function parseRunResultNodeIds(result: unknown): Set<string> {
  const ids = new Set<string>();
  if (!result || typeof result !== "object") return ids;
  const nodes = (result as { nodes?: unknown }).nodes;
  if (!Array.isArray(nodes)) return ids;
  for (const entry of nodes) {
    if (!entry || typeof entry !== "object") continue;
    const nodeId = (entry as { nodeId?: unknown }).nodeId;
    if (typeof nodeId === "string" && nodeId.trim()) {
      ids.add(nodeId);
    }
  }
  return ids;
}

export function nodeExecutionStatus(
  nodeId: string,
  run: Run | null | undefined,
): NodeExecutionStatus {
  if (!run) return "idle";

  if (run.currentWait?.nodeId === nodeId) {
    return "waiting";
  }

  if (parseRunResultNodeIds(run.result).has(nodeId)) {
    return "completed";
  }

  // Failed/cancelled runs clear result today — we cannot attribute failure to a
  // specific node without progress APIs. Keep chrome idle.
  return "idle";
}

export function nodeExecutionStatusLabel(
  status: NodeExecutionStatus,
): string | null {
  switch (status) {
    case "waiting":
      return "Waiting";
    case "completed":
      return "Done";
    case "failed":
      return "Failed";
    default:
      return null;
  }
}

/** Build a lookup for all canvas node ids against the active run. */
export function buildNodeExecutionStatusMap(
  nodeIds: readonly string[],
  run: Run | null | undefined,
): Record<string, NodeExecutionStatus> {
  const map: Record<string, NodeExecutionStatus> = {};
  for (const id of nodeIds) {
    map[id] = nodeExecutionStatus(id, run);
  }
  return map;
}
