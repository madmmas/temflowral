/**
 * Derive editable form fields from a node type's JSON Schema (`configSchema`
 * from GET /node-types). Kept pure for unit tests without a DOM (#91).
 * Field order prefers `x-order`, then known type overrides (#112).
 */

export type ConfigFieldKind =
  | "string"
  | "number"
  | "boolean"
  | "enum"
  | "json"
  | "stringMap";

export type ConfigField = {
  name: string;
  kind: ConfigFieldKind;
  required: boolean;
  enumValues?: string[];
  description?: string;
  minimum?: number;
  maximum?: number;
  /** Present when the property schema declared `x-order`. */
  order?: number;
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
  "x-order"?: unknown;
};

/** Authoring order when Go map JSON / missing x-order would scramble fields. */
export const NODE_TYPE_FIELD_ORDER: Record<string, readonly string[]> = {
  http: ["method", "url", "headers", "body"],
  wait: ["signal", "timeoutSeconds"],
  condition: ["field", "equals"],
  delay: ["seconds"],
  childWorkflow: ["graph", "input"],
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

function readXOrder(propertySchema: unknown): number | undefined {
  const property = asObject(propertySchema);
  if (!property) return undefined;
  const order = property["x-order"];
  return typeof order === "number" && Number.isFinite(order) ? order : undefined;
}

/** True when the property is an object of string values (HTTP headers). */
export function isStringMapPropertySchema(propertySchema: unknown): boolean {
  const schema = asObject(propertySchema) ?? {};
  if (primaryType(schema) !== "object") return false;
  const additional = schema.additionalProperties;
  if (additional === true) return false;
  if (typeof additional === "object" && additional !== null) {
    const valueSchema = asObject(additional);
    return primaryType(valueSchema ?? {}) === "string";
  }
  return false;
}

/**
 * Map a property schema to a form field kind.
 * Untyped / object / array properties become JSON textareas (e.g. condition
 * `equals`, childWorkflow `graph`), except string-maps → headers builder.
 */
export function fieldKindFromPropertySchema(
  propertySchema: unknown,
): ConfigFieldKind {
  if (isStringMapPropertySchema(propertySchema)) return "stringMap";

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

function sortConfigFields(
  fields: ConfigField[],
  nodeTypeId?: string,
): ConfigField[] {
  const hasXOrder = fields.some((field) => field.order !== undefined);
  if (hasXOrder) {
    return [...fields].sort((a, b) => {
      const ao = a.order ?? Number.POSITIVE_INFINITY;
      const bo = b.order ?? Number.POSITIVE_INFINITY;
      if (ao !== bo) return ao - bo;
      return a.name.localeCompare(b.name);
    });
  }

  const preferred = nodeTypeId
    ? NODE_TYPE_FIELD_ORDER[nodeTypeId]
    : undefined;
  if (preferred && preferred.length > 0) {
    const rank = new Map(preferred.map((name, index) => [name, index]));
    return [...fields].sort((a, b) => {
      const ar = rank.get(a.name) ?? preferred.length + 1;
      const br = rank.get(b.name) ?? preferred.length + 1;
      if (ar !== br) return ar - br;
      return a.name.localeCompare(b.name);
    });
  }

  return fields;
}

/**
 * List top-level properties of an object configSchema as form fields.
 * Schemas without `properties` yield an empty list (e.g. start / open noop).
 */
export function fieldsFromConfigSchema(
  configSchema: unknown,
  options?: { nodeTypeId?: string },
): ConfigField[] {
  const schema = asObject(configSchema);
  if (!schema) return [];

  const properties = schema.properties;
  if (!properties || typeof properties !== "object") return [];

  const required = new Set(
    Array.isArray(schema.required)
      ? schema.required.filter((name): name is string => typeof name === "string")
      : [],
  );

  const fields = Object.entries(properties).map(([name, propertySchema]) => {
    const property = asObject(propertySchema) ?? {};
    const kind = fieldKindFromPropertySchema(propertySchema);
    const order = readXOrder(propertySchema);
    const field: ConfigField = {
      name,
      kind,
      required: required.has(name),
      description:
        typeof property.description === "string"
          ? property.description
          : undefined,
    };
    if (order !== undefined) field.order = order;
    if (kind === "enum") {
      field.enumValues = stringEnumValues(property);
    }
    if (kind === "number") {
      if (typeof property.minimum === "number") field.minimum = property.minimum;
      if (typeof property.maximum === "number") field.maximum = property.maximum;
    }
    return field;
  });

  return sortConfigFields(fields, options?.nodeTypeId);
}

/** Format a config value for a text/JSON input. */
export function formatConfigValueForInput(
  kind: ConfigFieldKind,
  value: unknown,
): string {
  if (value === undefined || value === null) return "";
  if (kind === "json" || kind === "stringMap") {
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
    case "json":
    case "stringMap": {
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
