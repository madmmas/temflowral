"use client";

import {
  useCallback,
  useRef,
  useState,
  type ChangeEvent,
  type PointerEvent,
} from "react";

import {
  fieldsFromConfigSchema,
  formatConfigValueForInput,
  parseConfigInputValue,
  setConfigFieldValue,
  type ConfigField,
} from "@/lib/config-schema";
import { templateSuggestionsForNode } from "@/lib/config-templates";
import {
  supportsActivityOptions,
  type ActivityOptions,
  type CanvasEdge,
  type CanvasNode,
} from "@/lib/graph-canvas";
import {
  clampNodeConfigPanelWidth,
  NODE_CONFIG_PANEL_DEFAULT_WIDTH,
} from "@/lib/node-config-panel";
import type { NodeType } from "@/lib/node-types";
import { HeadersFieldEditor } from "@/components/headers-field-editor";
import { TemplateStringInput } from "@/components/template-string-input";

type NodeConfigPanelProps = {
  node: CanvasNode;
  nodeType: NodeType | undefined;
  nodes?: CanvasNode[];
  edges?: CanvasEdge[];
  width?: number;
  onWidthChange?: (width: number) => void;
  onChangeLabel: (label: string) => void;
  onChangeConfig: (config: Record<string, unknown>) => void;
  onChangeActivityOptions: (options: ActivityOptions | undefined) => void;
  onChangeTaskQueue: (taskQueue: string | undefined) => void;
  onClose: () => void;
};

/**
 * Side panel that edits the selected node's label and config from the
 * registry `configSchema` (#91), plus collapsed advanced activity fields (#93).
 * Width is user-resizable (#110).
 */
export function NodeConfigPanel({
  node,
  nodeType,
  nodes = [],
  edges = [],
  width = NODE_CONFIG_PANEL_DEFAULT_WIDTH,
  onWidthChange,
  onChangeLabel,
  onChangeConfig,
  onChangeActivityOptions,
  onChangeTaskQueue,
  onClose,
}: NodeConfigPanelProps) {
  const config = node.data.config ?? {};
  const fields = fieldsFromConfigSchema(nodeType?.configSchema, {
    nodeTypeId: node.data.nodeType,
  });
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});
  const showAdvanced = supportsActivityOptions(node.data.nodeType);
  const dragStartX = useRef<number | null>(null);
  const dragStartWidth = useRef(width);
  const panelWidth = clampNodeConfigPanelWidth(width);
  const templateSuggestions = templateSuggestionsForNode(
    nodes,
    edges,
    node.id,
  );

  const updateField = (field: ConfigField, raw: string) => {
    const parsed = parseConfigInputValue(field.kind, raw);
    if (!parsed.ok) {
      setFieldErrors((current) => ({ ...current, [field.name]: parsed.error }));
      return;
    }
    setFieldErrors((current) => {
      const next = { ...current };
      delete next[field.name];
      return next;
    });
    onChangeConfig(setConfigFieldValue(config, field.name, parsed.value));
  };

  const onResizePointerDown = useCallback(
    (event: PointerEvent<HTMLDivElement>) => {
      if (!onWidthChange) return;
      event.preventDefault();
      dragStartX.current = event.clientX;
      dragStartWidth.current = panelWidth;
      event.currentTarget.setPointerCapture(event.pointerId);
    },
    [onWidthChange, panelWidth],
  );

  const onResizePointerMove = useCallback(
    (event: PointerEvent<HTMLDivElement>) => {
      if (dragStartX.current === null || !onWidthChange) return;
      // Dragging the left handle leftward grows the panel.
      const delta = dragStartX.current - event.clientX;
      onWidthChange(
        clampNodeConfigPanelWidth(dragStartWidth.current + delta),
      );
    },
    [onWidthChange],
  );

  const onResizePointerUp = useCallback(
    (event: PointerEvent<HTMLDivElement>) => {
      if (dragStartX.current === null) return;
      dragStartX.current = null;
      if (event.currentTarget.hasPointerCapture(event.pointerId)) {
        event.currentTarget.releasePointerCapture(event.pointerId);
      }
    },
    [],
  );

  return (
    <aside
      data-testid="node-config-panel"
      className="relative flex h-full flex-col gap-3 overflow-y-auto border-l border-black/10 bg-white p-3 dark:border-white/15 dark:bg-neutral-950"
      style={{ width: panelWidth }}
    >
      {onWidthChange && (
        <div
          role="separator"
          aria-orientation="vertical"
          aria-label="Resize node config panel"
          data-testid="node-config-resize"
          onPointerDown={onResizePointerDown}
          onPointerMove={onResizePointerMove}
          onPointerUp={onResizePointerUp}
          onPointerCancel={onResizePointerUp}
          className="absolute inset-y-0 left-0 z-10 flex w-2 -translate-x-1/2 cursor-ew-resize touch-none items-center justify-center"
        >
          <span className="h-8 w-0.5 rounded-full bg-black/20 dark:bg-white/25" />
        </div>
      )}

      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <h2 className="text-xs font-semibold uppercase tracking-wide text-black/50 dark:text-white/50">
            Node config
          </h2>
          <p
            data-testid="node-config-editing"
            className="mt-0.5 truncate text-sm font-medium text-black/80 dark:text-white/80"
            title={node.data.label}
          >
            Editing: {node.data.label || node.data.nodeType}
          </p>
          <p className="mt-0.5 text-[10px] uppercase tracking-wide text-black/40 dark:text-white/40">
            {node.data.nodeType}
          </p>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="shrink-0 rounded px-1.5 py-0.5 text-xs text-black/50 hover:bg-black/5 dark:text-white/50 dark:hover:bg-white/10"
        >
          Close
        </button>
      </div>

      <label className="flex flex-col gap-1 text-xs">
        <span className="font-medium text-black/70 dark:text-white/70">
          Label
        </span>
        <input
          aria-label="Node label"
          value={node.data.label}
          onChange={(event) => onChangeLabel(event.target.value)}
          className="rounded-md border border-black/10 bg-white px-2 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
        />
      </label>

      {!nodeType && (
        <p className="text-xs text-black/50 dark:text-white/50">
          Unknown node type — config schema unavailable.
        </p>
      )}

      {nodeType && fields.length === 0 && (
        <p className="text-xs text-black/50 dark:text-white/50">
          This node type has no configurable fields.
        </p>
      )}

      {fields.map((field) => (
        <ConfigFieldInput
          key={field.name}
          field={field}
          value={config[field.name]}
          error={fieldErrors[field.name]}
          nodeId={node.id}
          templateSuggestions={templateSuggestions}
          onChange={(raw) => updateField(field, raw)}
          onChangeObject={(next) => {
            setFieldErrors((current) => {
              const copy = { ...current };
              delete copy[field.name];
              return copy;
            });
            onChangeConfig(setConfigFieldValue(config, field.name, next));
          }}
        />
      ))}

      {showAdvanced && (
        <AdvancedActivityFields
          activityOptions={node.data.activityOptions}
          taskQueue={node.data.taskQueue}
          onChangeActivityOptions={onChangeActivityOptions}
          onChangeTaskQueue={onChangeTaskQueue}
        />
      )}

      {nodeType?.description && (
        <p className="text-[11px] leading-snug text-black/45 dark:text-white/45">
          {nodeType.description}
        </p>
      )}
    </aside>
  );
}

function AdvancedActivityFields({
  activityOptions,
  taskQueue,
  onChangeActivityOptions,
  onChangeTaskQueue,
}: {
  activityOptions: ActivityOptions | undefined;
  taskQueue: string | undefined;
  onChangeActivityOptions: (options: ActivityOptions | undefined) => void;
  onChangeTaskQueue: (taskQueue: string | undefined) => void;
}) {
  const startToClose = activityOptions?.startToCloseTimeoutSeconds;
  const maxAttempts = activityOptions?.retryPolicy?.maximumAttempts;

  const patchOptions = (patch: {
    startToCloseTimeoutSeconds?: number | undefined;
    maximumAttempts?: number | undefined;
  }) => {
    const nextStart =
      patch.startToCloseTimeoutSeconds !== undefined
        ? patch.startToCloseTimeoutSeconds
        : startToClose;
    const nextAttempts =
      patch.maximumAttempts !== undefined ? patch.maximumAttempts : maxAttempts;

    if (nextStart === undefined && nextAttempts === undefined) {
      onChangeActivityOptions(undefined);
      return;
    }

    const next: ActivityOptions = {};
    if (nextStart !== undefined) {
      next.startToCloseTimeoutSeconds = nextStart;
    }
    if (nextAttempts !== undefined) {
      next.retryPolicy = { maximumAttempts: nextAttempts };
    }
    onChangeActivityOptions(next);
  };

  return (
    <details
      data-testid="advanced-activity-fields"
      className="rounded-md border border-black/10 bg-white/60 p-2 dark:border-white/15 dark:bg-neutral-900/60"
    >
      <summary className="cursor-pointer text-xs font-medium text-black/60 dark:text-white/60">
        Advanced (activity)
      </summary>
      <div className="mt-2 flex flex-col gap-2">
        <label className="flex flex-col gap-1 text-xs">
          <span className="font-medium text-black/70 dark:text-white/70">
            Task queue
          </span>
          <input
            aria-label="Task queue"
            value={taskQueue ?? ""}
            onChange={(event) => {
              const value = event.target.value.trim();
              onChangeTaskQueue(value || undefined);
            }}
            placeholder="temflowral"
            className="rounded-md border border-black/10 bg-white px-2 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs">
          <span className="font-medium text-black/70 dark:text-white/70">
            Start-to-close timeout (s)
          </span>
          <input
            aria-label="Start-to-close timeout seconds"
            type="number"
            min={0.001}
            max={86400}
            step="any"
            value={startToClose ?? ""}
            onChange={(event) => {
              const raw = event.target.value.trim();
              if (raw === "") {
                patchOptions({ startToCloseTimeoutSeconds: undefined });
                return;
              }
              const value = Number(raw);
              if (!Number.isFinite(value)) return;
              patchOptions({ startToCloseTimeoutSeconds: value });
            }}
            className="rounded-md border border-black/10 bg-white px-2 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
          />
        </label>
        <label className="flex flex-col gap-1 text-xs">
          <span className="font-medium text-black/70 dark:text-white/70">
            Max attempts
          </span>
          <input
            aria-label="Maximum attempts"
            type="number"
            min={1}
            max={100}
            step={1}
            value={maxAttempts ?? ""}
            onChange={(event) => {
              const raw = event.target.value.trim();
              if (raw === "") {
                patchOptions({ maximumAttempts: undefined });
                return;
              }
              const value = Number(raw);
              if (!Number.isFinite(value)) return;
              patchOptions({ maximumAttempts: Math.trunc(value) });
            }}
            className="rounded-md border border-black/10 bg-white px-2 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900"
          />
        </label>
        <p className="text-[10px] text-black/40 dark:text-white/40">
          Only valid on activity-backed nodes. Engine defaults apply when empty.
        </p>
      </div>
    </details>
  );
}

function ConfigFieldInput({
  field,
  value,
  error,
  nodeId,
  templateSuggestions,
  onChange,
  onChangeObject,
}: {
  field: ConfigField;
  value: unknown;
  error?: string;
  nodeId: string;
  templateSuggestions: readonly string[];
  onChange: (raw: string) => void;
  onChangeObject: (value: Record<string, string> | undefined) => void;
}) {
  const label = (
    <span className="font-medium text-black/70 dark:text-white/70">
      {field.name}
      {field.required ? " *" : ""}
    </span>
  );

  const commonClass =
    "rounded-md border border-black/10 bg-white px-2 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900";

  if (field.kind === "stringMap") {
    return (
      <HeadersFieldEditor
        name={field.name}
        value={value}
        syncKey={`${nodeId}:${field.name}`}
        required={field.required}
        onChange={onChangeObject}
      />
    );
  }

  if (field.kind === "enum" && field.enumValues) {
    return (
      <label className="flex flex-col gap-1 text-xs" data-testid={`config-field-${field.name}`}>
        {label}
        <select
          aria-label={field.name}
          value={typeof value === "string" ? value : ""}
          onChange={(event: ChangeEvent<HTMLSelectElement>) =>
            onChange(event.target.value)
          }
          className={commonClass}
        >
          <option value="">Select…</option>
          {field.enumValues.map((entry) => (
            <option key={entry} value={entry}>
              {entry}
            </option>
          ))}
        </select>
        {error && <span className="text-red-600 dark:text-red-400">{error}</span>}
      </label>
    );
  }

  if (field.kind === "boolean") {
    const checked = value === true;
    return (
      <label className="flex items-center gap-2 text-xs" data-testid={`config-field-${field.name}`}>
        <input
          aria-label={field.name}
          type="checkbox"
          checked={checked}
          onChange={(event) =>
            onChange(event.target.checked ? "true" : "false")
          }
        />
        {label}
        {error && <span className="text-red-600 dark:text-red-400">{error}</span>}
      </label>
    );
  }

  if (field.kind === "string") {
    return (
      <div data-testid={`config-field-${field.name}`}>
        <TemplateStringInput
          name={field.name}
          value={typeof value === "string" ? value : ""}
          multiline={field.name === "body"}
          required={field.required}
          suggestions={templateSuggestions}
          onChange={onChange}
        />
        {error && (
          <span className="text-xs text-red-600 dark:text-red-400">{error}</span>
        )}
      </div>
    );
  }

  if (field.kind === "json") {
    return (
      <label className="flex flex-col gap-1 text-xs" data-testid={`config-field-${field.name}`}>
        {label}
        <textarea
          aria-label={field.name}
          rows={6}
          defaultValue={formatConfigValueForInput(field.kind, value)}
          key={`${nodeId}:${field.name}:${formatConfigValueForInput(field.kind, value)}`}
          onBlur={(event) => onChange(event.target.value)}
          spellCheck={false}
          className={`${commonClass} font-mono text-xs`}
        />
        <span className="text-[10px] text-black/40 dark:text-white/40">
          JSON value — applied on blur
        </span>
        {error && <span className="text-red-600 dark:text-red-400">{error}</span>}
      </label>
    );
  }

  return (
    <label className="flex flex-col gap-1 text-xs" data-testid={`config-field-${field.name}`}>
      {label}
      <input
        aria-label={field.name}
        type={field.kind === "number" ? "number" : "text"}
        min={field.minimum}
        max={field.maximum}
        value={formatConfigValueForInput(field.kind, value)}
        onChange={(event) => onChange(event.target.value)}
        className={commonClass}
      />
      {error && <span className="text-red-600 dark:text-red-400">{error}</span>}
    </label>
  );
}
