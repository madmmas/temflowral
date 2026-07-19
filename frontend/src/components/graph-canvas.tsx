"use client";

import { useCallback } from "react";
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
} from "@xyflow/react";

import "@xyflow/react/dist/style.css";

import {
  createNode,
  type CanvasEdge,
  type CanvasNode,
} from "@/lib/graph-canvas";

const initialNodes: CanvasNode[] = [
  {
    id: "node-1",
    position: { x: 0, y: 0 },
    data: { label: "Node 1" },
  },
];

const initialEdges: CanvasEdge[] = [];

function GraphCanvasInner() {
  const [nodes, setNodes, onNodesChange] = useNodesState(initialNodes);
  const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
  const { screenToFlowPosition } = useReactFlow();

  const onConnect = useCallback(
    (connection: Connection) => {
      setEdges((current) => addEdge(connection, current));
    },
    [setEdges],
  );

  const onAddNode = useCallback(() => {
    // Drop the node near the current viewport centre.
    const position = screenToFlowPosition({
      x: window.innerWidth / 2,
      y: window.innerHeight / 2,
    });
    setNodes((current) => [...current, createNode(position)]);
  }, [screenToFlowPosition, setNodes]);

  return (
    <ReactFlow
      nodes={nodes}
      edges={edges}
      onNodesChange={onNodesChange}
      onEdgesChange={onEdgesChange}
      onConnect={onConnect}
      deleteKeyCode={["Backspace", "Delete"]}
      fitView
      proOptions={{ hideAttribution: true }}
    >
      <Panel position="top-left">
        <button
          type="button"
          onClick={onAddNode}
          className="rounded-md border border-black/10 bg-white px-3 py-1.5 text-sm font-medium shadow-sm hover:bg-black/5 dark:border-white/15 dark:bg-neutral-900 dark:hover:bg-white/10"
        >
          Add node
        </button>
      </Panel>
      <Panel position="top-right">
        <p className="rounded-md bg-white/70 px-2 py-1 text-xs text-black/60 dark:bg-neutral-900/70 dark:text-white/60">
          Drag handles to connect · select + Delete to remove
        </p>
      </Panel>
      <Background />
      <MiniMap pannable zoomable />
      <Controls />
    </ReactFlow>
  );
}

/** Base workflow canvas: pan/zoom, add nodes, connect, and delete. */
export function GraphCanvas() {
  return (
    <ReactFlowProvider>
      <GraphCanvasInner />
    </ReactFlowProvider>
  );
}
