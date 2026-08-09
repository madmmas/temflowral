"use client";

import { createContext, useContext, type ReactNode } from "react";

import type { Run } from "@/lib/node-execution-status";
import {
  buildNodeExecutionStatusMap,
  type NodeExecutionStatus,
} from "@/lib/node-execution-status";

const NodeExecutionStatusContext = createContext<
  Record<string, NodeExecutionStatus>
>({});

type NodeExecutionStatusProviderProps = {
  run: Run | null;
  nodeIds: readonly string[];
  children: ReactNode;
};

/** Supplies per-node execution chrome for WorkflowNode (#113). */
export function NodeExecutionStatusProvider({
  run,
  nodeIds,
  children,
}: NodeExecutionStatusProviderProps) {
  const statusByNodeId = buildNodeExecutionStatusMap(nodeIds, run);
  return (
    <NodeExecutionStatusContext.Provider value={statusByNodeId}>
      {children}
    </NodeExecutionStatusContext.Provider>
  );
}

export function useNodeExecutionStatus(
  nodeId: string,
): NodeExecutionStatus {
  const map = useContext(NodeExecutionStatusContext);
  return map[nodeId] ?? "idle";
}
