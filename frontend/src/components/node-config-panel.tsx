"use client";

import { useState, type ChangeEvent } from "react";

import {
  fieldsFromConfigSchema,
  formatConfigValueForInput,
  parseConfigInputValue,
  setConfigFieldValue,
  type ConfigField,
} from "@/lib/config-schema";
import type { CanvasNode } from "@/lib/graph-canvas";
import type { NodeType } from "@/lib/node-types";

type NodeConfigPanelProps = {
  node: CanvasNode;
  nodeType: NodeType | undefined;
  onChangeLabel: (label: string) => void;
  onChangeConfig: (config: Record<string, unknown>) => void;
  onClose: () => void;
};

/**
 * Side panel that edits the selected node's label and config from the
 * registry `configSchema` (#91).
 */
export function NodeConfigPanel({
  node,
  nodeType,
  onChangeLabel,
  onChangeConfig,
  onClose,
}: NodeConfigPanelProps) {
  const config = node.data.config ?? {};
  const fields = fieldsFromConfigSchema(nodeType?.configSchema);
  const [fieldErrors, setFieldErrors] = useState<Record<string, string>>({});

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

  return (
    <aside
      data-testid="node-config-panel"
      className="flex w-72 shrink-0 flex-col gap-3 overflow-y-auto border-l border-black/10 bg-black/[0.02] p-3 dark:border-white/15 dark:bg-white/[0.03]"
    >
      <div className="flex items-start justify-between gap-2">
        <div>
          <h2 className="text-xs font-semibold uppercase tracking-wide text-black/50 dark:text-white/50">
            Node config
          </h2>
          <p className="mt-0.5 text-[10px] uppercase tracking-wide text-black/40 dark:text-white/40">
            {node.data.nodeType}
          </p>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="rounded px-1.5 py-0.5 text-xs text-black/50 hover:bg-black/5 dark:text-white/50 dark:hover:bg-white/10"
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
          onChange={(raw) => updateField(field, raw)}
        />
      ))}

      {nodeType?.description && (
        <p className="text-[11px] leading-snug text-black/45 dark:text-white/45">
          {nodeType.description}
        </p>
      )}
    </aside>
  );
}

function ConfigFieldInput({
  field,
  value,
  error,
  onChange,
}: {
  field: ConfigField;
  value: unknown;
  error?: string;
  onChange: (raw: string) => void;
}) {
  const label = (
    <span className="font-medium text-black/70 dark:text-white/70">
      {field.name}
      {field.required ? " *" : ""}
    </span>
  );

  const commonClass =
    "rounded-md border border-black/10 bg-white px-2 py-1.5 text-sm outline-none focus:ring-2 focus:ring-blue-500/40 dark:border-white/15 dark:bg-neutral-900";

  if (field.kind === "enum" && field.enumValues) {
    return (
      <label className="flex flex-col gap-1 text-xs">
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
      <label className="flex items-center gap-2 text-xs">
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

  if (field.kind === "json" || (field.kind === "string" && field.name === "body")) {
    return (
      <label className="flex flex-col gap-1 text-xs">
        {label}
        <textarea
          aria-label={field.name}
          rows={field.kind === "json" ? 6 : 3}
          defaultValue={formatConfigValueForInput(field.kind, value)}
          key={`${field.name}:${formatConfigValueForInput(field.kind, value)}`}
          onBlur={(event) => onChange(event.target.value)}
          spellCheck={false}
          className={`${commonClass} font-mono text-xs`}
        />
        {field.kind === "json" && (
          <span className="text-[10px] text-black/40 dark:text-white/40">
            JSON value — applied on blur
          </span>
        )}
        {error && <span className="text-red-600 dark:text-red-400">{error}</span>}
      </label>
    );
  }

  return (
    <label className="flex flex-col gap-1 text-xs">
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
