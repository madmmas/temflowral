"use client";

import { useEffect, useState } from "react";

import { createApiClient } from "@/api";
import type { components } from "@/api";

export type NodeType = components["schemas"]["NodeType"];

const DEFAULT_CATEGORY = "other";

/**
 * Group node types by their `category`, preserving input order within each
 * group. Types without a category fall under `other`.
 *
 * Pure helper so palette grouping is unit-testable without a DOM or network.
 */
export function groupByCategory(
  nodeTypes: readonly NodeType[],
): [string, NodeType[]][] {
  const groups = new Map<string, NodeType[]>();
  for (const nodeType of nodeTypes) {
    const category = nodeType.category?.trim() || DEFAULT_CATEGORY;
    const existing = groups.get(category);
    if (existing) {
      existing.push(nodeType);
    } else {
      groups.set(category, [nodeType]);
    }
  }
  return [...groups.entries()];
}

export type NodeTypesState = {
  nodeTypes: NodeType[];
  loading: boolean;
  error: string | null;
};

/**
 * Fetch the node-type registry from `GET /node-types` via the typed client.
 * The palette renders from this rather than a hardcoded list (#16), so a new
 * backend node type appears without a frontend change.
 */
export function useNodeTypes(): NodeTypesState {
  const [state, setState] = useState<NodeTypesState>({
    nodeTypes: [],
    loading: true,
    error: null,
  });

  useEffect(() => {
    let cancelled = false;
    const client = createApiClient();

    client
      .GET("/node-types")
      .then(({ data, error }) => {
        if (cancelled) return;
        if (error || !data) {
          setState({
            nodeTypes: [],
            loading: false,
            error: "Failed to load node types",
          });
          return;
        }
        setState({ nodeTypes: data.nodeTypes, loading: false, error: null });
      })
      .catch(() => {
        if (cancelled) return;
        setState({
          nodeTypes: [],
          loading: false,
          error: "Failed to load node types",
        });
      });

    return () => {
      cancelled = true;
    };
  }, []);

  return state;
}
