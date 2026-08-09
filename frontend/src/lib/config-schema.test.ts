import { describe, expect, it } from "vitest";

import {
  fieldKindFromPropertySchema,
  fieldsFromConfigSchema,
  formatConfigValueForInput,
  parseConfigInputValue,
  setConfigFieldValue,
} from "./config-schema";

const httpSchema = {
  type: "object",
  required: ["method", "url"],
  additionalProperties: false,
  properties: {
    // Deliberately scrambled entry order (Go map / alpha) — form order must
    // still be method → url → headers → body via nodeTypeId override / x-order.
    body: { type: "string" },
    headers: {
      type: "object",
      additionalProperties: { type: "string" },
    },
    method: {
      type: "string",
      enum: ["GET", "POST", "PUT", "PATCH", "DELETE"],
    },
    url: { type: "string", maxLength: 2048 },
  },
};

const delaySchema = {
  type: "object",
  required: ["seconds"],
  properties: {
    seconds: { type: "number", minimum: 0, maximum: 604800 },
  },
};

const conditionSchema = {
  type: "object",
  required: ["field", "equals"],
  properties: {
    field: { type: "string", minLength: 1 },
    equals: {},
  },
};

const waitSchema = {
  type: "object",
  required: ["signal", "timeoutSeconds"],
  properties: {
    signal: { type: "string" },
    timeoutSeconds: { type: "number", minimum: 0 },
  },
};

describe("fieldsFromConfigSchema", () => {
  it("maps HTTP configSchema fields including enums and string-map headers", () => {
    const fields = fieldsFromConfigSchema(httpSchema, { nodeTypeId: "http" });
    expect(fields.map((field) => [field.name, field.kind, field.required])).toEqual([
      ["method", "enum", true],
      ["url", "string", true],
      ["headers", "stringMap", false],
      ["body", "string", false],
    ]);
    expect(fields[0]?.enumValues).toEqual([
      "GET",
      "POST",
      "PUT",
      "PATCH",
      "DELETE",
    ]);
  });

  it("orders by x-order when present", () => {
    const fields = fieldsFromConfigSchema({
      type: "object",
      properties: {
        body: { type: "string", "x-order": 4 },
        method: { type: "string", "x-order": 1 },
        url: { type: "string", "x-order": 2 },
        headers: {
          type: "object",
          "x-order": 3,
          additionalProperties: { type: "string" },
        },
      },
    });
    expect(fields.map((field) => field.name)).toEqual([
      "method",
      "url",
      "headers",
      "body",
    ]);
  });

  it("maps delay seconds as a required number", () => {
    const fields = fieldsFromConfigSchema(delaySchema);
    expect(fields).toEqual([
      {
        name: "seconds",
        kind: "number",
        required: true,
        description: undefined,
        minimum: 0,
        maximum: 604800,
      },
    ]);
  });

  it("maps condition equals (empty schema) as JSON", () => {
    const fields = fieldsFromConfigSchema(conditionSchema);
    expect(fields).toEqual([
      {
        name: "field",
        kind: "string",
        required: true,
        description: undefined,
      },
      {
        name: "equals",
        kind: "json",
        required: true,
        description: undefined,
      },
    ]);
  });

  it("maps wait signal and timeoutSeconds", () => {
    const fields = fieldsFromConfigSchema(waitSchema);
    expect(fields.map((field) => field.name)).toEqual([
      "signal",
      "timeoutSeconds",
    ]);
    expect(fields[1]?.kind).toBe("number");
  });

  it("returns no fields when properties are absent", () => {
    expect(
      fieldsFromConfigSchema({ type: "object", additionalProperties: false }),
    ).toEqual([]);
  });
});

describe("fieldKindFromPropertySchema", () => {
  it("classifies common JSON Schema fragments", () => {
    expect(fieldKindFromPropertySchema({ type: "string" })).toBe("string");
    expect(fieldKindFromPropertySchema({ type: "integer" })).toBe("number");
    expect(fieldKindFromPropertySchema({ type: "boolean" })).toBe("boolean");
    expect(
      fieldKindFromPropertySchema({ type: "string", enum: ["a", "b"] }),
    ).toBe("enum");
    expect(fieldKindFromPropertySchema({ type: "object" })).toBe("json");
    expect(
      fieldKindFromPropertySchema({
        type: "object",
        additionalProperties: { type: "string" },
      }),
    ).toBe("stringMap");
    expect(fieldKindFromPropertySchema({})).toBe("json");
  });
});

describe("config value parse/format", () => {
  it("parses numbers and JSON", () => {
    expect(parseConfigInputValue("number", "1.5")).toEqual({
      ok: true,
      value: 1.5,
    });
    expect(parseConfigInputValue("json", '{"a":1}')).toEqual({
      ok: true,
      value: { a: 1 },
    });
    expect(parseConfigInputValue("json", "{").ok).toBe(false);
  });

  it("formats values for inputs", () => {
    expect(formatConfigValueForInput("string", "hi")).toBe("hi");
    expect(formatConfigValueForInput("json", { a: 1 })).toBe(
      JSON.stringify({ a: 1 }, null, 2),
    );
    expect(formatConfigValueForInput("number", undefined)).toBe("");
  });

  it("sets and clears config keys", () => {
    expect(setConfigFieldValue({ a: 1 }, "b", 2)).toEqual({ a: 1, b: 2 });
    expect(setConfigFieldValue({ a: 1, b: 2 }, "b", undefined)).toEqual({
      a: 1,
    });
  });
});
