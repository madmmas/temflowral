"use client";

import { useCallback, useEffect, useState } from "react";
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
} from "@xyflow/react";

import "@xyflow/react/dist/style.css";

import { createApiClient, type components } from "@/api";
import { NODE_TYPE_DRAG_KEY, NodePalette } from "@/components/node-palette";
import { WorkflowNode } from "@/components/nodes/workflow-node";
import {
  apiErrorMessage,
  createNode,
  isTerminalRunStatus,
  serializeGraph,
  type CanvasEdge,
  type CanvasNode,
} from "@/lib/graph-canvas";
import type { NodeType } from "@/lib/node-types";

const initialNodes: CanvasNode[] = [];
const initialEdges: CanvasEdge[] = [];
const RUN_POLL_INTERVAL_MS = 1_500;

const nodeTypes: NodeTypes = {
  workflow: WorkflowNode,
};

type Run = components["schemas"]["Run"];
type Action = "idle" | "saving" | "starting";

function GraphCanvasInner() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const [graphName, setGraphName] = useState("Untitled workflow");
  const [savedGraphId, setSavedGraphId] = useState<string | null>(null);
  const [run, setRun] = useState<Run | null>(null);
  const [action, setAction] = useState<Action>("idle");
  const [actionError, setActionError] = useState<string | null>(null);
  const { screenToFlowPosition } = useReactFlow();
  const runId = run?.id;
  const runStatus = run?.status;

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

  // Click a palette item: drop near the viewport centre.
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
    try {
      const { data, error } = await createApiClient().POST("/graphs", {
        body: serializeGraph(graphName, nodes, edges),
      });
      if (error || !data) {
        setActionError(apiErrorMessage(error, "Failed to save graph"));
        return null;
      }
      setSavedGraphId(data.id);
      return data;
    } catch {
      setActionError("Failed to save graph");
      return null;
    } finally {
      setAction("idle");
    }
  }, [edges, graphName, nodes]);

  const runGraph = useCallback(async () => {
    const graph = await saveGraph();
    if (!graph) return;

    setAction("starting");
    setActionError(null);
    setRun(null);
    try {
      const { data, error } = await createApiClient().POST(
        "/graphs/{graphId}/run",
        {
          params: { path: { graphId: graph.id } },
          body: {},
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
  }, [saveGraph]);

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();
      const typeId = event.dataTransfer.getData(NODE_TYPE_DRAG_KEY);
      if (!typeId) return;
      addNodeAt(
        { id: typeId, name: typeId, configSchema: {} },
        event.clientX,
        event.clientY,
      );
    },
    [addNodeAt],
  );

  return (
    <div data-testid="graph-editor" className="flex h-full w-full">
      <NodePalette onAddNodeType={onAddNodeType} />
      <div className="flex min-w-0 flex-1 flex-col">
        <div className="flex flex-wrap items-center gap-2 border-b border-black/10 px-3 py-2 dark:border-white/15">
          <input
            aria-label="Graph name"
            value={graphName}
            onChange={(event) => setGraphName(event.target.value)}
            className="min-w-40 flex-1 rounded-md border border-black/10 bg-white px-2.5 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
          />
          <button
            type="button"
            onClick={saveGraph}
            disabled={action !== "idle" || nodes.length === 0}
            className="rounded-md border border-black/10 bg-white px-3 py-1.5 text-sm font-medium shadow-sm hover:bg-black/5 disabled:cursor-not-allowed disabled:opacity-50 dark:border-white/15 dark:bg-neutral-900 dark:hover:bg-white/10"
          >
            {action === "saving" ? "Saving…" : "Save"}
          </button>
          <button
            type="button"
            onClick={runGraph}
            disabled={action !== "idle" || nodes.length === 0}
            className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white shadow-sm hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {action === "starting" ? "Starting…" : "Run"}
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
            onDragOver={onDragOver}
            onDrop={onDrop}
            deleteKeyCode={["Backspace", "Delete"]}
            fitView
            proOptions={{ hideAttribution: true }}
          >
            <Panel position="top-right">
              <p className="rounded-md bg-white/70 px-2 py-1 text-xs text-black/60 dark:bg-neutral-900/70 dark:text-white/60">
                Drag or click a node · drag handles to connect · select + Delete
                to remove
              </p>
            </Panel>
            <Background />
            <MiniMap pannable zoomable />
            <Controls />
          </ReactFlow>
        </div>
        {(savedGraphId || run || actionError) && (
          <div
            role={actionError ? "alert" : "status"}
            className="flex min-h-9 flex-wrap items-center gap-3 border-t border-black/10 px-3 py-2 text-xs dark:border-white/15"
          >
            {savedGraphId && (
              <span className="text-black/50 dark:text-white/50">
                Graph saved: {savedGraphId}
              </span>
            )}
            {run && (
              <span
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
            {run?.error && (
              <span className="text-red-600 dark:text-red-400">{run.error}</span>
            )}
            {actionError && (
              <span className="text-red-600 dark:text-red-400">
                {actionError}
              </span>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

/** Workflow canvas with registry palette and save/run integration (#15–#17). */
export function GraphCanvas() {
  return (
    <ReactFlowProvider>
      <GraphCanvasInner />
    </ReactFlowProvider>
  );
}
