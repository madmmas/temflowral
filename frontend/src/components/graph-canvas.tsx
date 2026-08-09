"use client";

import {
  useCallback,
  useEffect,
  useState,
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
import { AuthoringTip, EmptyCanvasGuide } from "@/components/canvas-guidance";
import { CanvasErrorBanner } from "@/components/canvas-error-banner";
import { GraphIdChip } from "@/components/graph-id-chip";
import { NODE_TYPE_DRAG_KEY, NodePalette } from "@/components/node-palette";
import { NodeConfigPanel } from "@/components/node-config-panel";
import { WorkflowNode } from "@/components/nodes/workflow-node";
import { RunHistoryList } from "@/components/run-history-list";
import { RunResultPanel } from "@/components/run-result-panel";
import { RunSignalPanel } from "@/components/run-signal-panel";
import { RunStatusChip } from "@/components/run-status-chip";
import { WorkflowLibrary } from "@/components/workflow-library";
import {
  isCanvasBannerDismissed,
  pickCanvasBannerMessage,
} from "@/lib/canvas-error-banner";
import {
  paletteClickScreenOffset,
  readAuthoringTipDismissed,
  writeAuthoringTipDismissed,
} from "@/lib/canvas-guidance";
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
import {
  DEFAULT_GRAPH_NAME,
  findGraphsWithSameName,
  graphNameHint,
  resolveGraphNameForSave,
} from "@/lib/graph-naming";
import { NodeTypeRegistryProvider } from "@/lib/node-type-registry";
import { useNodeTypes, type NodeType } from "@/lib/node-types";
import {
  historyEntryToRun,
  runsForGraph,
  runToHistoryEntry,
  upsertRunHistory,
  type RunHistoryEntry,
} from "@/lib/run-history";
import {
  runHasResultPanelContent,
  RUN_RESULT_PANEL_DEFAULT_HEIGHT,
} from "@/lib/run-result-panel";
import {
  clampNodeConfigPanelWidth,
  NODE_CONFIG_PANEL_DEFAULT_WIDTH,
  readNodeConfigPanelWidth,
  writeNodeConfigPanelWidth,
} from "@/lib/node-config-panel";
import {
  MINIMAP_STYLE,
  MINIMAP_VISIBLE_DEFAULT,
  minimapColorsForScheme,
  readMinimapVisible,
  writeMinimapVisible,
} from "@/lib/minimap-prefs";
import { graphShortId } from "@/lib/workflow-library";

const initialNodes: CanvasNode[] = [];
const initialEdges: CanvasEdge[] = [];
const RUN_POLL_INTERVAL_MS = 1_500;
const GRAPH_QUERY_PARAM = "graph";
const EMPTY_GRAPH_FINGERPRINT = graphFingerprint(
  DEFAULT_GRAPH_NAME,
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
  const [graphName, setGraphName] = useState(DEFAULT_GRAPH_NAME);
  const [savedGraphId, setSavedGraphId] = useState<string | null>(null);
  const [baselineFingerprint, setBaselineFingerprint] = useState(
    EMPTY_GRAPH_FINGERPRINT,
  );
  const [run, setRun] = useState<Run | null>(null);
  const [runHistory, setRunHistory] = useState<RunHistoryEntry[]>([]);
  const [action, setAction] = useState<Action>("idle");
  const [actionError, setActionError] = useState<string | null>(null);
  const [bannerDismissed, setBannerDismissed] = useState<{
    source: "action" | "graphs";
    message: string;
  } | null>(null);
  const [selectedNodeId, setSelectedNodeId] = useState<string | null>(null);
  const [resultPanelVisible, setResultPanelVisible] = useState(false);
  const [resultPanelCollapsed, setResultPanelCollapsed] = useState(false);
  const [resultPanelHeight, setResultPanelHeight] = useState(
    RUN_RESULT_PANEL_DEFAULT_HEIGHT,
  );
  const [configPanelWidth, setConfigPanelWidth] = useState(
    NODE_CONFIG_PANEL_DEFAULT_WIDTH,
  );
  const [minimapVisible, setMinimapVisible] = useState(MINIMAP_VISIBLE_DEFAULT);
  const [minimapPrefersDark, setMinimapPrefersDark] = useState(true);
  const [authoringTipDismissed, setAuthoringTipDismissed] = useState(false);
  const [libraryOpen, setLibraryOpen] = useState(false);
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
  const hasResultContent = runHasResultPanelContent(run);
  const showResultPanel = resultPanelVisible && hasResultContent;
  const graphRunHistory = runsForGraph(runHistory, savedGraphId);
  const duplicateNameCount = findGraphsWithSameName(
    graphs,
    graphName,
    savedGraphId,
  ).length;
  const nameHint = graphNameHint({
    name: graphName,
    isNewGraph: savedGraphId === null,
    duplicateCount: duplicateNameCount,
  });
  const banner = pickCanvasBannerMessage(actionError, graphsError);
  const showErrorBanner =
    banner !== null &&
    !(
      banner.source === "graphs" &&
      isCanvasBannerDismissed(bannerDismissed, banner)
    );
  const dirty =
    graphFingerprint(graphName, nodes, edges) !== baselineFingerprint;

  useEffect(() => {
    setMinimapVisible(readMinimapVisible());
    setAuthoringTipDismissed(readAuthoringTipDismissed());
    setConfigPanelWidth(readNodeConfigPanelWidth());
    if (typeof window === "undefined" || !window.matchMedia) return;
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const sync = () => setMinimapPrefersDark(media.matches);
    sync();
    media.addEventListener("change", sync);
    return () => media.removeEventListener("change", sync);
  }, []);

  useEffect(() => {
    if (!run) return;
    setRunHistory((prev) => upsertRunHistory(prev, runToHistoryEntry(run)));
  }, [run]);

  useEffect(() => {
    if (!hasResultContent) return;
    setResultPanelVisible(true);
    setResultPanelCollapsed(false);
  }, [runId, hasResultContent]);
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
      setResultPanelVisible(false);
      setActionError(null);
      setBannerDismissed(null);
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
          // Invalid deep links should not stick in the URL after a failed open.
          replaceGraphQuery(null);
          return;
        }
        applyLoadedGraph(data);
        refreshGraphs();
      } catch {
        setActionError("Failed to open graph");
        replaceGraphQuery(null);
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
      const offset = paletteClickScreenOffset(nodes.length);
      addNodeAt(
        nodeType,
        window.innerWidth / 2 + offset.x,
        window.innerHeight / 2 + offset.y,
      );
    },
    [addNodeAt, nodes.length],
  );

  const saveGraph = useCallback(async () => {
    if (nodes.length === 0) {
      setActionError("Add at least one node before saving");
      return null;
    }

    const resolved = resolveGraphNameForSave({
      name: graphName,
      isNewGraph: savedGraphId === null,
      graphs,
      excludeGraphId: savedGraphId,
      prompt: (message, defaultValue) =>
        typeof window === "undefined"
          ? null
          : window.prompt(message, defaultValue ?? ""),
      confirm: (message) =>
        typeof window === "undefined" ? true : window.confirm(message),
    });
    if (!resolved.ok) {
      if (resolved.reason === "placeholder" && resolved.message) {
        setActionError(resolved.message);
      }
      return null;
    }
    if (resolved.name !== graphName) {
      setGraphName(resolved.name);
    }

    setAction("saving");
    setActionError(null);
    const body = serializeGraph(resolved.name, nodes, edges);
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
      setBaselineFingerprint(graphFingerprint(resolved.name, nodes, edges));
      refreshGraphs();
      return data;
    } catch {
      setActionError("Failed to save graph");
      return null;
    } finally {
      setAction("idle");
    }
  }, [edges, graphName, graphs, nodes, refreshGraphs, savedGraphId]);

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
      setGraphName(DEFAULT_GRAPH_NAME);
      setNodes([]);
      setEdges([]);
      setSavedGraphId(null);
      setBaselineFingerprint(EMPTY_GRAPH_FINGERPRINT);
      setSelectedNodeId(null);
      setRun(null);
      setResultPanelVisible(false);
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

  const onOpenGraphFromLibrary = useCallback(
    (graphId: string) => {
      if (graphId === savedGraphId) {
        setLibraryOpen(false);
        return;
      }
      if (!confirmDiscardIfDirty(dirty)) return;
      setLibraryOpen(false);
      void loadGraphById(graphId);
    },
    [dirty, loadGraphById, savedGraphId],
  );

  const onNewGraph = useCallback(() => {
    if (!confirmDiscardIfDirty(dirty)) return;
    setGraphName(DEFAULT_GRAPH_NAME);
    setNodes([]);
    setEdges([]);
    setSavedGraphId(null);
    setBaselineFingerprint(EMPTY_GRAPH_FINGERPRINT);
    setSelectedNodeId(null);
    setRun(null);
    setResultPanelVisible(false);
    setActionError(null);
    setBannerDismissed(null);
    replaceGraphQuery(null);
  }, [dirty, setEdges, setNodes]);

  const onDismissErrorBanner = useCallback(() => {
    if (!banner) return;
    if (banner.source === "action") {
      setActionError(null);
      return;
    }
    setBannerDismissed(banner);
  }, [banner]);

  const onSelectHistoryRun = useCallback((entry: RunHistoryEntry) => {
    setRun(historyEntryToRun(entry));
    setActionError(null);
    if (runHasResultPanelContent(entry)) {
      setResultPanelVisible(true);
      setResultPanelCollapsed(false);
    }
  }, []);

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
      <div className="relative flex min-w-0 flex-1 flex-col">
        {showErrorBanner && banner && (
          <CanvasErrorBanner
            message={banner.message}
            onDismiss={onDismissErrorBanner}
            onNew={onNewGraph}
          />
        )}
        <div
          data-testid="graph-toolbar"
          className="flex flex-nowrap items-center gap-2 overflow-x-auto border-b border-black/10 px-3 py-2 dark:border-white/15"
        >
          <div className="flex min-w-0 flex-1 items-center gap-2">
            <div className="flex min-w-0 w-full max-w-md flex-col gap-0.5">
              <input
                aria-label="Graph name"
                value={graphName}
                onChange={(event) => setGraphName(event.target.value)}
                className="min-w-32 w-full rounded-md border border-black/10 bg-white px-2.5 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
              />
              {nameHint && (
                <span
                  data-testid="graph-name-hint"
                  className="truncate text-[11px] text-amber-800 dark:text-amber-200"
                >
                  {nameHint}
                </span>
              )}
            </div>
            {dirty ? (
              <span
                data-testid="unsaved-indicator"
                className="shrink-0 rounded-full bg-amber-500/15 px-2 py-0.5 text-[11px] font-medium text-amber-800 dark:bg-amber-400/15 dark:text-amber-200"
              >
                Unsaved
              </span>
            ) : savedGraphId ? (
              <span
                data-testid="saved-indicator"
                className="shrink-0 text-[11px] font-medium text-black/40 dark:text-white/40"
              >
                Saved
              </span>
            ) : null}
          </div>

          <div className="flex shrink-0 items-center gap-2">
            <button
              type="button"
              data-testid="open-graph"
              aria-label="Open workflow library"
              onClick={() => setLibraryOpen(true)}
              disabled={busy}
              className="max-w-44 truncate rounded-md border border-black/10 bg-white px-3 py-1.5 text-sm font-medium shadow-sm hover:bg-black/5 disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/15 dark:bg-neutral-900 dark:hover:bg-white/10"
            >
              {savedGraphId
                ? `Open… (${graphShortId(savedGraphId)})`
                : graphsLoading
                  ? "Loading…"
                  : "Open…"}
            </button>
            <button
              type="button"
              onClick={onNewGraph}
              disabled={busy}
              className="rounded-md border border-black/10 bg-white px-3 py-1.5 text-sm font-medium shadow-sm hover:bg-black/5 disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/15 dark:bg-neutral-900 dark:hover:bg-white/10"
            >
              New
            </button>
          </div>

          <div
            data-testid="primary-actions"
            className="flex shrink-0 items-center gap-2"
          >
            <button
              type="button"
              onClick={saveGraph}
              disabled={busy || nodes.length === 0 || !dirty}
              className="rounded-md border border-black/10 bg-white px-3 py-1.5 text-sm font-medium shadow-sm hover:bg-black/5 disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/15 dark:bg-neutral-900 dark:hover:bg-white/10"
            >
              {action === "saving"
                ? "Saving…"
                : !dirty && savedGraphId
                  ? "Saved"
                  : "Save"}
            </button>
            <button
              type="button"
              onClick={runGraph}
              disabled={busy || nodes.length === 0}
              className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
            >
              {action === "starting" ? "Starting…" : "Run"}
            </button>
          </div>

          <div
            data-testid="danger-actions"
            className="ml-1 flex shrink-0 items-center border-l border-black/10 pl-3 dark:border-white/15"
          >
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
        </div>
        <div data-testid="graph-canvas" className="relative min-h-0 flex-1">
          <EmptyCanvasGuide
            visible={nodes.length === 0 && action !== "loading"}
          />
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
              <div className="flex flex-col items-end gap-1.5">
                <AuthoringTip
                  visible={nodes.length > 0 && !authoringTipDismissed}
                  onDismiss={() => {
                    setAuthoringTipDismissed(true);
                    writeAuthoringTipDismissed(true);
                  }}
                />
                <button
                  type="button"
                  data-testid="toggle-minimap"
                  aria-pressed={minimapVisible}
                  onClick={() => {
                    setMinimapVisible((current) => {
                      const next = !current;
                      writeMinimapVisible(next);
                      return next;
                    });
                  }}
                  className="rounded-md border border-black/10 bg-white/90 px-2 py-1 text-[11px] font-medium text-black/70 shadow-sm hover:bg-black/5 dark:border-white/15 dark:bg-neutral-900/90 dark:text-white/70 dark:hover:bg-white/10"
                >
                  {minimapVisible ? "Hide map" : "Show map"}
                </button>
              </div>
            </Panel>
            <Background />
            {minimapVisible && (
              <MiniMap
                pannable
                zoomable
                style={MINIMAP_STYLE}
                {...minimapColorsForScheme(minimapPrefersDark)}
                className="!rounded-md !border !border-black/20 !shadow-sm dark:!border-white/20"
              />
            )}
            <Controls />
          </ReactFlow>
          {selectedNode && (
            <div className="pointer-events-none absolute inset-y-0 right-0 z-20 flex">
              <div className="pointer-events-auto h-full shadow-lg">
                <NodeConfigPanel
                  node={selectedNode}
                  nodeType={selectedNodeType}
                  width={configPanelWidth}
                  onWidthChange={(next) => {
                    const clamped = clampNodeConfigPanelWidth(next);
                    setConfigPanelWidth(clamped);
                    writeNodeConfigPanelWidth(clamped);
                  }}
                  onChangeLabel={onChangeSelectedLabel}
                  onChangeConfig={onChangeSelectedConfig}
                  onChangeActivityOptions={onChangeSelectedActivityOptions}
                  onChangeTaskQueue={onChangeSelectedTaskQueue}
                  onClose={() => setSelectedNodeId(null)}
                />
              </div>
            </div>
          )}
        </div>
        {run?.currentWait && runId && (
          <RunSignalPanel
            runId={runId}
            currentWait={run.currentWait}
            busy={busy}
            onSend={sendRunSignal}
          />
        )}
        {showResultPanel && run && (
          <RunResultPanel
            result={run.result}
            error={run.error}
            collapsed={resultPanelCollapsed}
            height={resultPanelHeight}
            onCollapsedChange={setResultPanelCollapsed}
            onHeightChange={setResultPanelHeight}
            onDismiss={() => setResultPanelVisible(false)}
          />
        )}
        {(savedGraphId ||
          run ||
          graphRunHistory.length > 0 ||
          action === "loading") && (
          <div
            role="status"
            className="flex min-h-9 flex-wrap items-center gap-3 border-t border-black/10 px-3 py-2 text-xs dark:border-white/15"
          >
            {action === "loading" && (
              <span className="text-black/50 dark:text-white/50">
                Opening graph…
              </span>
            )}
            {savedGraphId && <GraphIdChip graphId={savedGraphId} />}
            {run && (
              <RunStatusChip
                status={run.status}
                waiting={Boolean(run.currentWait)}
                canOpenResult={hasResultContent}
                resultPanelOpen={showResultPanel}
                onOpenResult={() => {
                  setResultPanelVisible(true);
                  setResultPanelCollapsed(false);
                }}
              />
            )}
            <RunHistoryList
              entries={graphRunHistory}
              activeRunId={runId ?? null}
              onSelect={onSelectHistoryRun}
            />
          </div>
        )}
      </div>
      <WorkflowLibrary
        open={libraryOpen}
        graphs={graphs}
        loading={graphsLoading}
        error={graphsError}
        currentGraphId={savedGraphId}
        busy={busy}
        onClose={() => setLibraryOpen(false)}
        onOpenGraph={onOpenGraphFromLibrary}
        onRefresh={refreshGraphs}
      />
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
