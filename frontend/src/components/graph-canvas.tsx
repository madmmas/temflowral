"use client";

import {
  useCallback,
  useEffect,
  useState,
  type ChangeEvent,
} from "react";
import {
  addEdge,
  Background,
  Controls,
  MiniMap,
  Panel,
  ReactFlow,
  ReactFlowProvider,
  useEdgesState,
  useNodesState,
  useReactFlow,
  type Connection,
  type NodeTypes,
  type OnSelectionChangeParams,
} from "@xyflow/react";

import "@xyflow/react/dist/style.css";

import { createApiClient, type components } from "@/api";
import { NODE_TYPE_DRAG_KEY, NodePalette } from "@/components/node-palette";
import { NodeConfigPanel } from "@/components/node-config-panel";
import { WorkflowNode } from "@/components/nodes/workflow-node";
import { RunSignalPanel } from "@/components/run-signal-panel";
import {
  apiErrorMessage,
  createNode,
  deserializeGraph,
  graphFingerprint,
  isTerminalRunStatus,
  serializeGraph,
  type ActivityOptions,
  type CanvasEdge,
  type CanvasNode,
} from "@/lib/graph-canvas";
import { useGraphList } from "@/lib/graph-list";
import { NodeTypeRegistryProvider } from "@/lib/node-type-registry";
import { useNodeTypes, type NodeType } from "@/lib/node-types";

const initialNodes: CanvasNode[] = [];
const initialEdges: CanvasEdge[] = [];
const RUN_POLL_INTERVAL_MS = 1_500;
const GRAPH_QUERY_PARAM = "graph";
const EMPTY_GRAPH_FINGERPRINT = graphFingerprint(
  "Untitled workflow",
  initialNodes,
  initialEdges,
);

const nodeTypes: NodeTypes = {
  workflow: WorkflowNode,
};

type Run = components["schemas"]["Run"];
type Action = "idle" | "saving" | "starting" | "loading" | "deleting";

function confirmDiscardIfDirty(dirty: boolean): boolean {
  if (!dirty) return true;
  if (typeof window === "undefined") return true;
  return window.confirm(
    "You have unsaved changes. Discard them and continue?",
  );
}
function readGraphIdFromLocation(): string | null {
  if (typeof window === "undefined") return null;
  const value = new URLSearchParams(window.location.search).get(
    GRAPH_QUERY_PARAM,
  );
  return value?.trim() || null;
}

function replaceGraphQuery(graphId: string | null): void {
  if (typeof window === "undefined") return;
  const url = new URL(window.location.href);
  if (graphId) {
    url.searchParams.set(GRAPH_QUERY_PARAM, graphId);
  } else {
    url.searchParams.delete(GRAPH_QUERY_PARAM);
  }
  window.history.replaceState({}, "", `${url.pathname}${url.search}${url.hash}`);
}

function GraphCanvasInner() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [graphName, setGraphName] = useState("Untitled workflow");
  const [savedGraphId, setSavedGraphId] = useState<string | null>(null);
  const [baselineFingerprint, setBaselineFingerprint] = useState(
    EMPTY_GRAPH_FINGERPRINT,
  );
  const [run, setRun] = useState<Run | null>(null);
  const [action, setAction] = useState<Action>("idle");
  const [actionError, setActionError] = useState<string | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const { screenToFlowPosition } = useReactFlow();
  const {
    nodeTypes: registryNodeTypes,
    loading: nodeTypesLoading,
    error: nodeTypesError,
  } = useNodeTypes();
  const {
    graphs,
    loading: graphsLoading,
    error: graphsError,
    refresh: refreshGraphs,
  } = useGraphList();
  const runId = run?.id;
  const runStatus = run?.status;
  const dirty =
    graphFingerprint(graphName, nodes, edges) !== baselineFingerprint;
  const selectedNode =
    selectedNodeId === null
      ? undefined
      : nodes.find((node) => node.id === selectedNodeId);
  const selectedNodeType = selectedNode
    ? registryNodeTypes.find((entry) => entry.id === selectedNode.data.nodeType)
    : undefined;

  const applyLoadedGraph = useCallback(
    (graph: components["schemas"]["Graph"]) => {
      const deserialized = deserializeGraph(graph);
      setGraphName(deserialized.name);
      setNodes(deserialized.nodes);
      setEdges(deserialized.edges);
      setSavedGraphId(graph.id);
      setBaselineFingerprint(
        graphFingerprint(
          deserialized.name,
          deserialized.nodes,
          deserialized.edges,
        ),
      );
      setSelectedNodeId(null);
      setRun(null);
      setActionError(null);
      replaceGraphQuery(graph.id);
    },
    [setEdges, setNodes],
  );

  const loadGraphById = useCallback(
    async (graphId: string) => {
      setAction("loading");
      setActionError(null);
      try {
        const { data, error } = await createApiClient().GET("/graphs/{graphId}", {
          params: { path: { graphId } },
        });
        if (error || !data) {
          setActionError(apiErrorMessage(error, "Failed to open graph"));
          return;
        }
        applyLoadedGraph(data);
        refreshGraphs();
      } catch {
        setActionError("Failed to open graph");
      } finally {
        setAction("idle");
      }
    },
    [applyLoadedGraph, refreshGraphs],
  );

  useEffect(() => {
    const graphId = readGraphIdFromLocation();
    if (!graphId) return;
    void loadGraphById(graphId);
    // Open once from the initial URL; subsequent opens go through the picker.
    // eslint-disable-next-line react-hooks/exhaustive-deps -- mount-only hydrate
  }, []);

  useEffect(() => {
    if (!dirty) return;
    const onBeforeUnload = (event: BeforeUnloadEvent) => {
      event.preventDefault();
      event.returnValue = "";
    };
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [dirty]);

  useEffect(() => {
    if (!runId || !runStatus || isTerminalRunStatus(runStatus)) return;

    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;
    const client = createApiClient();

    const poll = async () => {
      try {
        const { data, error } = await client.GET("/runs/{runId}", {
          params: { path: { runId } },
        });
        if (cancelled) return;
        if (error || !data) {
          setActionError(apiErrorMessage(error, "Failed to refresh run status"));
          return;
        }
        setRun(data);
        if (!isTerminalRunStatus(data.status)) {
          timer = setTimeout(poll, RUN_POLL_INTERVAL_MS);
        }
      } catch {
        if (!cancelled) setActionError("Failed to refresh run status");
      }
    };

    timer = setTimeout(poll, RUN_POLL_INTERVAL_MS);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [runId, runStatus]);

  const onSelectionChange = useCallback(
    ({ nodes: selected }: OnSelectionChangeParams) => {
      setSelectedNodeId(selected.length === 1 ? selected[0].id : null);
    },
    [],
  );

  const onChangeSelectedLabel = useCallback(
    (label: string) => {
      if (!selectedNodeId) return;
      setNodes((current) =>
        current.map((node) =>
          node.id === selectedNodeId
            ? { ...node, data: { ...node.data, label } }
            : node,
        ),
      );
    },
    [selectedNodeId, setNodes],
  );

  const onChangeSelectedConfig = useCallback(
    (config: Record<string, unknown>) => {
      if (!selectedNodeId) return;
      setNodes((current) =>
        current.map((node) =>
          node.id === selectedNodeId
            ? { ...node, data: { ...node.data, config } }
            : node,
        ),
      );
    },
    [selectedNodeId, setNodes],
  );

  const onChangeSelectedActivityOptions = useCallback(
    (activityOptions: ActivityOptions | undefined) => {
      if (!selectedNodeId) return;
      setNodes((current) =>
        current.map((node) =>
          node.id === selectedNodeId
            ? { ...node, data: { ...node.data, activityOptions } }
            : node,
        ),
      );
    },
    [selectedNodeId, setNodes],
  );

  const onChangeSelectedTaskQueue = useCallback(
    (taskQueue: string | undefined) => {
      if (!selectedNodeId) return;
      setNodes((current) =>
        current.map((node) =>
          node.id === selectedNodeId
            ? { ...node, data: { ...node.data, taskQueue } }
            : node,
        ),
      );
    },
    [selectedNodeId, setNodes],
  );

  const onConnect = useCallback(
    (connection: Connection) => {
      setEdges((current) => addEdge(connection, current));
    },
    [setEdges],
  );

  const addNodeAt = useCallback(
    (nodeType: NodeType, screenX: number, screenY: number) => {
      const position = screenToFlowPosition({ x: screenX, y: screenY });
      setNodes((current) => [
        ...current,
        createNode(position, {
          nodeType: nodeType.id,
          label: nodeType.name,
          category: nodeType.category,
        }),
      ]);
    },
    [screenToFlowPosition, setNodes],
  );

  const onAddNodeType = useCallback(
    (nodeType: NodeType) => {
      addNodeAt(nodeType, window.innerWidth / 2, window.innerHeight / 2);
    },
    [addNodeAt],
  );

  const saveGraph = useCallback(async () => {
    if (nodes.length === 0) {
      setActionError("Add at least one node before saving");
      return null;
    }

    setAction("saving");
    setActionError(null);
    const body = serializeGraph(graphName, nodes, edges);
    const client = createApiClient();
    try {
      const result = savedGraphId
        ? await client.PUT("/graphs/{graphId}", {
            params: { path: { graphId: savedGraphId } },
            body,
          })
        : await client.POST("/graphs", { body });

      const { data, error } = result;
      if (error || !data) {
        setActionError(apiErrorMessage(error, "Failed to save graph"));
        return null;
      }
      setSavedGraphId(data.id);
      replaceGraphQuery(data.id);
      setBaselineFingerprint(graphFingerprint(graphName, nodes, edges));
      refreshGraphs();
      return data;
    } catch {
      setActionError("Failed to save graph");
      return null;
    } finally {
      setAction("idle");
    }
  }, [edges, graphName, nodes, refreshGraphs, savedGraphId]);

  const runGraph = useCallback(async () => {
    let graphId = savedGraphId;
    if (dirty || !graphId) {
      const graph = await saveGraph();
      if (!graph) return;
      graphId = graph.id;
    }

    setAction("starting");
    setActionError(null);
    setRun(null);
    const idempotencyKey = crypto.randomUUID();
    try {
      const { data, error } = await createApiClient().POST(
        "/graphs/{graphId}/run",
        {
          params: { path: { graphId } },
          body: { idempotencyKey },
        },
      );
      if (error || !data) {
        setActionError(apiErrorMessage(error, "Failed to start graph run"));
        return;
      }
      setRun(data);
    } catch {
      setActionError("Failed to start graph run");
    } finally {
      setAction("idle");
    }
  }, [dirty, saveGraph, savedGraphId]);

  const deleteGraph = useCallback(async () => {
    if (!savedGraphId) return;
    if (
      typeof window !== "undefined" &&
      !window.confirm("Delete this saved graph? This cannot be undone.")
    ) {
      return;
    }
    setAction("deleting");
    setActionError(null);
    try {
      const { error } = await createApiClient().DELETE("/graphs/{graphId}", {
        params: { path: { graphId: savedGraphId } },
      });
      if (error) {
        setActionError(apiErrorMessage(error, "Failed to delete graph"));
        return;
      }
      setGraphName("Untitled workflow");
      setNodes([]);
      setEdges([]);
      setSavedGraphId(null);
      setBaselineFingerprint(EMPTY_GRAPH_FINGERPRINT);
      setSelectedNodeId(null);
      setRun(null);
      replaceGraphQuery(null);
      refreshGraphs();
    } catch {
      setActionError("Failed to delete graph");
    } finally {
      setAction("idle");
    }
  }, [refreshGraphs, savedGraphId, setEdges, setNodes]);

  const sendRunSignal = useCallback(
    async (signal: string, payload: unknown | undefined) => {
      if (!runId) {
        throw new Error("No active run");
      }
      const body: components["schemas"]["SignalRunRequest"] =
        payload === undefined ? { signal } : { signal, payload };
      const { data, error } = await createApiClient().POST(
        "/runs/{runId}/signal",
        {
          params: { path: { runId } },
          body,
        },
      );
      if (error || !data) {
        throw new Error(apiErrorMessage(error, "Failed to send signal"));
      }
      // Resume polling with a fresh GET so currentWait clears promptly.
      const refreshed = await createApiClient().GET("/runs/{runId}", {
        params: { path: { runId } },
      });
      if (refreshed.data) {
        setRun(refreshed.data);
      }
    },
    [runId],
  );

  const onOpenGraphChange = useCallback(
    (event: ChangeEvent<HTMLSelectElement>) => {
      const graphId = event.target.value;
      if (!graphId) {
        if (!confirmDiscardIfDirty(dirty)) {
          event.target.value = savedGraphId ?? "";
          return;
        }
        setSavedGraphId(null);
        replaceGraphQuery(null);
        return;
      }
      if (!confirmDiscardIfDirty(dirty)) {
        event.target.value = savedGraphId ?? "";
        return;
      }
      void loadGraphById(graphId);
    },
    [dirty, loadGraphById, savedGraphId],
  );

  const onNewGraph = useCallback(() => {
    if (!confirmDiscardIfDirty(dirty)) return;
    setGraphName("Untitled workflow");
    setNodes([]);
    setEdges([]);
    setSavedGraphId(null);
    setBaselineFingerprint(EMPTY_GRAPH_FINGERPRINT);
    setSelectedNodeId(null);
    setRun(null);
    setActionError(null);
    replaceGraphQuery(null);
  }, [dirty, setEdges, setNodes]);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();
      const typeId = event.dataTransfer.getData(NODE_TYPE_DRAG_KEY);
      if (!typeId) return;
      const registryType = registryNodeTypes.find((entry) => entry.id === typeId);
      addNodeAt(
        registryType ?? { id: typeId, name: typeId, configSchema: {} },
        event.clientX,
        event.clientY,
      );
    },
    [addNodeAt, registryNodeTypes],
  );

  const busy = action !== "idle";

  return (
    <NodeTypeRegistryProvider nodeTypes={registryNodeTypes}>
    <div data-testid="graph-editor" className="flex h-full w-full">
      <NodePalette
        onAddNodeType={onAddNodeType}
        nodeTypes={registryNodeTypes}
        loading={nodeTypesLoading}
        error={nodeTypesError}
      />
      <div className="flex min-w-0 flex-1 flex-col">
        <div className="flex flex-wrap items-center gap-2 border-b border-black/10 px-3 py-2 dark:border-white/15">
          <input
            aria-label="Graph name"
            value={graphName}
            onChange={(event) => setGraphName(event.target.value)}
            className="min-w-40 flex-1 rounded-md border border-black/10 bg-white px-2.5 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
          />
          <select
            aria-label="Open saved graph"
            data-testid="open-graph"
            value={savedGraphId ?? ""}
            onChange={onOpenGraphChange}
            disabled={busy || graphsLoading}
            className="max-w-56 rounded-md border border-black/10 bg-white px-2 py-1.5 text-sm dark:border-white/15 dark:bg-neutral-900"
          >
            <option value="">
              {graphsLoading ? "Loading graphs…" : "Open saved graph…"}
            </option>
            {graphs.map((graph) => (
              <option key={graph.id} value={graph.id}>
                {graph.name?.trim() || "Untitled"} ({graph.id.slice(0, 8)})
              </option>
            ))}
          </select>
          <button
            type="button"
            onClick={onNewGraph}
            disabled={busy}
            className="rounded-md border border-black/10 bg-white px-3 py-1.5 text-sm font-medium shadow-sm hover:bg-black/5 disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/15 dark:bg-neutral-900 dark:hover:bg-white/10"
          >
            New
          </button>
          <button
            type="button"
            onClick={saveGraph}
            disabled={busy || nodes.length === 0 || !dirty}
            className="rounded-md border border-black/10 bg-white px-3 py-1.5 text-sm font-medium shadow-sm hover:bg-black/5 disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/15 dark:bg-neutral-900 dark:hover:bg-white/10"
          >
            {action === "saving" ? "Saving…" : dirty ? "Save" : "Saved"}
          </button>
          <button
            type="button"
            onClick={runGraph}
            disabled={busy || nodes.length === 0}
            className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {action === "starting" ? "Starting…" : "Run"}
          </button>
          <button
            type="button"
            onClick={() => void deleteGraph()}
            disabled={busy || !savedGraphId}
            data-testid="delete-graph"
            className="rounded-md border border-red-200 bg-white px-3 py-1.5 text-sm font-medium text-red-700 shadow-sm hover:bg-red-50 disabled:cursor-not-allowed disabled:opacity-50 dark:border-red-900/50 dark:bg-neutral-900 dark:text-red-400 dark:hover:bg-red-950/40"
          >
            {action === "deleting" ? "Deleting…" : "Delete"}
          </button>
        </div>
        <div data-testid="graph-canvas" className="relative min-h-0 flex-1">
          <ReactFlow
            nodes={nodes}
            edges={edges}
            nodeTypes={nodeTypes}
            onNodesChange={onNodesChange}
            onEdgesChange={onEdgesChange}
            onConnect={onConnect}
            onSelectionChange={onSelectionChange}
            onDragOver={onDragOver}
            onDrop={onDrop}
            deleteKeyCode={["Backspace", "Delete"]}
            fitView
            proOptions={{ hideAttribution: true }}
          >
            <Panel position="top-right">
              <p className="rounded-md bg-white/70 px-2 py-1 text-xs text-black/60 dark:bg-neutral-900/70 dark:text-white/60">
                Select a node to edit config · drag handles to connect · Delete
                to remove
              </p>
            </Panel>
            <Background />
            <MiniMap pannable zoomable />
            <Controls />
          </ReactFlow>
        </div>
        {(savedGraphId || run || actionError || graphsError || action === "loading") && (
          <div
            role={actionError || graphsError ? "alert" : "status"}
            className="flex min-h-9 flex-wrap items-center gap-3 border-t border-black/10 px-3 py-2 text-xs dark:border-white/15"
          >
            {action === "loading" && (
              <span className="text-black/50 dark:text-white/50">
                Opening graph…
              </span>
            )}
            {savedGraphId && (
              <span className="text-black/50 dark:text-white/50">
                Graph: {savedGraphId}
              </span>
            )}
            {run && (
              <span
                data-testid="run-status"
                className={
                  run.status === "failed" || run.status === "cancelled"
                    ? "font-medium text-red-600 dark:text-red-400"
                    : run.status === "completed"
                      ? "font-medium text-green-600 dark:text-green-400"
                      : "font-medium text-blue-600 dark:text-blue-400"
                }
              >
                Run {run.status}
              </span>
            )}
            {run?.currentWait && runId && (
              <RunSignalPanel
                runId={runId}
                currentWait={run.currentWait}
                busy={busy}
                onSend={sendRunSignal}
              />
            )}
            {run?.result && (
              <pre
                data-testid="run-result"
                className="max-w-full overflow-x-auto rounded bg-black/5 px-2 py-1 text-black/70 dark:bg-white/10 dark:text-white/70"
              >
                {JSON.stringify(run.result, null, 2)}
              </pre>
            )}
            {run?.error && (
              <span className="text-red-600 dark:text-red-400">{run.error}</span>
            )}
            {graphsError && (
              <span className="text-red-600 dark:text-red-400">
                {graphsError}
              </span>
            )}
            {actionError && (
              <span className="text-red-600 dark:text-red-400">
                {actionError}
              </span>
            )}
          </div>
        )}
      </div>
      {selectedNode && (
        <NodeConfigPanel
          node={selectedNode}
          nodeType={selectedNodeType}
          onChangeLabel={onChangeSelectedLabel}
          onChangeConfig={onChangeSelectedConfig}
          onChangeActivityOptions={onChangeSelectedActivityOptions}
          onChangeTaskQueue={onChangeSelectedTaskQueue}
          onClose={() => setSelectedNodeId(null)}
        />
      )}
    </div>
    </NodeTypeRegistryProvider>
  );
}

/** Workflow canvas with registry palette and save/run/reopen integration. */
export function GraphCanvas() {
  return (
    <ReactFlowProvider>
      <GraphCanvasInner />
    </ReactFlowProvider>
  );
}
