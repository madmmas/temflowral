import { readFileSync } from "node:fs";
import path from "node:path";
import Ajv, { type ErrorObject, type ValidateFunction } from "ajv";
import addFormats from "ajv-formats";
import { load } from "js-yaml";

/**
 * Contract conformance uses api/openapi.yaml as the single source of truth:
 * we validate real (or mock) HTTP responses against the exact component
 * schemas the spec declares, so drift between the implementation and the
 * contract is caught here rather than downstream in the frontend.
 */
const SPEC_PATH = path.resolve(__dirname, "../../api/openapi.yaml");

interface OpenApiDocument {
  components?: {
    schemas?: Record<string, unknown>;
  };
}

function loadContractSchemas(): Record<string, unknown> {
  const doc = load(readFileSync(SPEC_PATH, "utf8")) as OpenApiDocument;
  const schemas = doc.components?.schemas;
  if (!schemas) {
    throw new Error(`No components.schemas found in ${SPEC_PATH}`);
  }
  return schemas;
}

// strict:false so OpenAPI-only keywords (example, description on the wrapper)
// and unknown formats (e.g. "double") are ignored rather than throwing.
const ajv = new Ajv({ strict: false, allErrors: true });
addFormats(ajv);

// Register the schemas under a single root so intra-spec $refs like
// "#/components/schemas/Node" resolve without hand-dereferencing.
ajv.addSchema(
  { $id: "contract", components: { schemas: loadContractSchemas() } },
  "contract",
);

export function validatorFor(schemaName: string): ValidateFunction {
  const validate = ajv.getSchema(
    `contract#/components/schemas/${schemaName}`,
  );
  if (!validate) {
    throw new Error(`Schema not found in contract: ${schemaName}`);
  }
  return validate;
}

export function formatErrors(
  schemaName: string,
  errors: ErrorObject[] | null | undefined,
): string {
  if (!errors || errors.length === 0) {
    return `Response did not match schema ${schemaName}`;
  }
  const details = errors
    .map((err) => `  - ${err.instancePath || "<root>"} ${err.message ?? ""}`)
    .join("\n");
  return `Response did not match schema ${schemaName}:\n${details}`;
}
