"use client";

import { createContext, useContext, type ReactNode } from "react";

import type { NodeType } from "@/lib/node-types";

const NodeTypeRegistryContext = createContext<readonly NodeType[]>([]);

/** Provide registry node types to React Flow custom node renderers. */
export function NodeTypeRegistryProvider({
  nodeTypes,
  children,
}: {
  nodeTypes: readonly NodeType[];
  children: ReactNode;
}) {
  return (
    <NodeTypeRegistryContext.Provider value={nodeTypes}>
      {children}
    </NodeTypeRegistryContext.Provider>
  );
}

export function useNodeTypeRegistry(): readonly NodeType[] {
  return useContext(NodeTypeRegistryContext);
}
