"use client";

import { useCallback, useEffect, useState } from "react";

import type { GraphSummary } from "@/lib/graph-canvas";
import { apiErrorMessage } from "@/lib/graph-canvas";
import { useCreateApiClient } from "@/lib/api-client-context";

export type GraphListState = {
  graphs: GraphSummary[];
  loading: boolean;
  error: string | null;
  refresh: () => void;
};

/**
 * Load saved graph summaries from `GET /graphs` for the reopen picker (#90).
 */
export function useGraphList(enabled = true): GraphListState {
  const createClient = useCreateApiClient();
  const [graphs, setGraphs] = useState<GraphSummary[]>([]);
  const [loading, setLoading] = useState(enabled);
  const [error, setError] = useState<string | null>(null);
  const [tick, setTick] = useState(0);

  const refresh = useCallback(() => {
    setTick((value) => value + 1);
  }, []);

  useEffect(() => {
    if (!enabled) {
      setLoading(false);
      return;
    }

    let cancelled = false;
    setLoading(true);
    const client = createClient();

    client
      .GET("/graphs")
      .then(({ data, error: apiError }) => {
        if (cancelled) return;
        if (apiError || !data) {
          setGraphs([]);
          setError(apiErrorMessage(apiError, "Failed to load graphs"));
          setLoading(false);
          return;
        }
        setGraphs(data.graphs);
        setError(null);
        setLoading(false);
      })
      .catch(() => {
        if (cancelled) return;
        setGraphs([]);
        setError("Failed to load graphs");
        setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [enabled, tick, createClient]);

  return { graphs, loading, error, refresh };
}
