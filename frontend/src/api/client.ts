import createClient from "openapi-fetch";

import { getApiBaseUrl } from "@/lib/api-base-url";

import type { components, paths } from "./generated/schema";

export type { components, paths };

/** Typed OpenAPI client. Paths/schemas come from `src/api/generated/schema.ts`. */
export type ApiClient = ReturnType<typeof createApiClient>;

/**
 * Build a typed API client.
 *
 * Defaults to `NEXT_PUBLIC_API_BASE_URL`, falling back to the Prism mock on
 * `:4010` (see CONTRIBUTING.md). Pass an explicit base URL in tests.
 */
export function createApiClient(baseUrl: string = getApiBaseUrl()) {
  return createClient<paths>({ baseUrl });
}
