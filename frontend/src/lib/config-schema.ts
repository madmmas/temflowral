/**
 * Derive editable form fields from a node type's JSON Schema (`configSchema`
 * from GET /node-types). Kept pure for unit tests without a DOM (#91).
 */

export type ConfigFieldKind = "string" | "number" | "boolean" | "enum" | "json";

export type ConfigField = {
  name: string;
  kind: ConfigFieldKind;
  required: boolean;
  enumValues?: string[];
  description?: string;
  minimum?: number;
  maximum?: number;
};

type JsonSchemaObject = {
  type?: string | string[];
  properties?: Record<string, unknown>;
  required?: string[];
  enum?: unknown[];
  description?: string;
  minimum?: number;
  maximum?: number;
  additionalProperties?: unknown;
};

function asObject(value: unknown): JsonSchemaObject | null {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    return null;
  }
  return value as JsonSchemaObject;
}

function primaryType(schema: JsonSchemaObject): string | undefined {
  if (typeof schema.type === "string") return schema.type;
  if (Array.isArray(schema.type)) {
    return schema.type.find((entry) => entry !== "null");
  }
  return undefined;
}

function stringEnumValues(schema: JsonSchemaObject): string[] | undefined {
  if (!Array.isArray(schema.enum) || schema.enum.length === 0) return undefined;
  if (!schema.enum.every((entry) => typeof entry === "string")) return undefined;
  return schema.enum as string[];
}

/**
 * Map a property schema to a form field kind.
 * Untyped / object / array properties become JSON textareas (e.g. condition
 * `equals`, HTTP `headers`, childWorkflow `graph`).
 */
export function fieldKindFromPropertySchema(
  propertySchema: unknown,
): ConfigFieldKind {
  const schema = asObject(propertySchema) ?? {};
  const enumValues = stringEnumValues(schema);
  if (enumValues) return "enum";

  const type = primaryType(schema);
  switch (type) {
    case "string":
      return "string";
    case "number":
    case "integer":
      return "number";
    case "boolean":
      return "boolean";
    case "object":
    case "array":
      return "json";
    default:
      // Empty schema `{}` (condition.equals) or unknown → JSON.
      return "json";
  }
}

/**
 * List top-level properties of an object configSchema as form fields.
 * Schemas without `properties` yield an empty list (e.g. start / open noop).
 */
export function fieldsFromConfigSchema(configSchema: unknown): ConfigField[] {
  const schema = asObject(configSchema);
  if (!schema) return [];

  const properties = schema.properties;
  if (!properties || typeof properties !== "object") return [];

  const required = new Set(
    Array.isArray(schema.required)
      ? schema.required.filter((name): name is string => typeof name === "string")
      : [],
  );

  return Object.entries(properties).map(([name, propertySchema]) => {
    const property = asObject(propertySchema) ?? {};
    const kind = fieldKindFromPropertySchema(propertySchema);
    const field: ConfigField = {
      name,
      kind,
      required: required.has(name),
      description:
        typeof property.description === "string"
          ? property.description
          : undefined,
    };
    if (kind === "enum") {
      field.enumValues = stringEnumValues(property);
    }
    if (kind === "number") {
      if (typeof property.minimum === "number") field.minimum = property.minimum;
      if (typeof property.maximum === "number") field.maximum = property.maximum;
    }
    return field;
  });
}

/** Format a config value for a text/JSON input. */
export function formatConfigValueForInput(
  kind: ConfigFieldKind,
  value: unknown,
): string {
  if (value === undefined || value === null) return "";
  if (kind === "json") {
    try {
      return JSON.stringify(value, null, 2);
    } catch {
      return String(value);
    }
  }
  if (kind === "boolean") return value ? "true" : "false";
  return String(value);
}

/**
 * Parse a form string into a config value. Returns `{ ok: false }` when JSON
 * fields contain invalid JSON (caller should keep the previous value / show error).
 */
export function parseConfigInputValue(
  kind: ConfigFieldKind,
  raw: string,
): { ok: true; value: unknown } | { ok: false; error: string } {
  const trimmed = raw.trim();

  switch (kind) {
    case "string":
    case "enum":
      return { ok: true, value: raw };
    case "number": {
      if (trimmed === "") return { ok: true, value: undefined };
      const value = Number(trimmed);
      if (!Number.isFinite(value)) {
        return { ok: false, error: "Must be a number" };
      }
      return { ok: true, value };
    }
    case "boolean":
      if (trimmed === "") return { ok: true, value: undefined };
      if (trimmed === "true") return { ok: true, value: true };
      if (trimmed === "false") return { ok: true, value: false };
      return { ok: false, error: "Must be true or false" };
    case "json": {
      if (trimmed === "") return { ok: true, value: undefined };
      try {
        return { ok: true, value: JSON.parse(trimmed) as unknown };
      } catch {
        return { ok: false, error: "Invalid JSON" };
      }
    }
    default:
      return { ok: true, value: raw };
  }
}

/**
 * Apply a field update onto a config object. Omits keys when the new value is
 * `undefined` (cleared optional fields).
 */
export function setConfigFieldValue(
  config: Record<string, unknown>,
  field: string,
  value: unknown,
): Record<string, unknown> {
  const next = { ...config };
  if (value === undefined) {
    delete next[field];
  } else {
    next[field] = value;
  }
  return next;
}
