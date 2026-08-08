import type { components } from "@/api";

export type NodeType = components["schemas"]["NodeType"];
export type NodeOutputHandle = components["schemas"]["NodeOutputHandle"];

/**
 * Resolve the source handles to render for a canvas node from registry metadata.
 * Fixed `outputHandles` win; otherwise derive IDs from `outputHandlesFromConfig`
 * (mirrors backend `nodetype.ResolveOutputHandles` for the shapes we support).
 * Empty means a single unnamed default source handle.
 */
export function resolveOutputHandles(
  nodeType: NodeType | undefined,
  config: Record<string, unknown> | undefined,
): NodeOutputHandle[] {
  if (!nodeType) return [];

  const fixed = nodeType.outputHandles;
  if (fixed && fixed.length > 0) {
    return fixed.map((handle) => ({
      id: handle.id,
      label: handle.label,
    }));
  }

  const fromConfig = nodeType.outputHandlesFromConfig;
  if (!fromConfig?.path) return [];

  const value = lookupPath(config ?? {}, fromConfig.path);
  if (value === undefined) return [];

  const ids = handleIDsFromValue(value);
  return ids.map((id) => ({ id }));
}

function lookupPath(
  root: Record<string, unknown>,
  path: string,
): unknown | undefined {
  const parts = path.split(".").filter(Boolean);
  let current: unknown = root;
  for (const part of parts) {
    if (
      typeof current !== "object" ||
      current === null ||
      Array.isArray(current)
    ) {
      return undefined;
    }
    current = (current as Record<string, unknown>)[part];
  }
  return current;
}

function handleIDsFromValue(value: unknown): string[] {
  if (Array.isArray(value)) {
    const ids: string[] = [];
    for (const item of value) {
      if (typeof item === "string" && item) {
        ids.push(item);
        continue;
      }
      if (
        typeof item === "object" &&
        item !== null &&
        "id" in item &&
        typeof (item as { id: unknown }).id === "string" &&
        (item as { id: string }).id
      ) {
        ids.push((item as { id: string }).id);
      }
    }
    return ids;
  }
  if (typeof value === "object" && value !== null && !Array.isArray(value)) {
    return Object.keys(value as Record<string, unknown>).sort();
  }
  return [];
}
