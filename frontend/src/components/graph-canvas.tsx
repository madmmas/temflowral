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
  type NodeTypes,
} from "@xyflow/react";

import "@xyflow/react/dist/style.css";

import { NODE_TYPE_DRAG_KEY, NodePalette } from "@/components/node-palette";
import { WorkflowNode } from "@/components/nodes/workflow-node";
import {
  createNode,
  type CanvasEdge,
  type CanvasNode,
} from "@/lib/graph-canvas";
import type { NodeType } from "@/lib/node-types";

const initialNodes: CanvasNode[] = [];
const initialEdges: CanvasEdge[] = [];

const nodeTypes: NodeTypes = {
  workflow: WorkflowNode,
};

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
    <div className="flex h-full w-full">
      <NodePalette onAddNodeType={onAddNodeType} />
      <div className="min-w-0 flex-1">
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
              Drag or click a node · drag handles to connect · select + Delete to
              remove
            </p>
          </Panel>
          <Background />
          <MiniMap pannable zoomable />
          <Controls />
        </ReactFlow>
      </div>
    </div>
  );
}

/** Base workflow canvas with a node-type-driven palette (#15, #16). */
export function GraphCanvas() {
  return (
    <ReactFlowProvider>
      <GraphCanvasInner />
    </ReactFlowProvider>
  );
}
