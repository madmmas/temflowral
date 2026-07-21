import createClient from "openapi-fetch";

import { getApiBaseUrl } from "@/lib/api-base-url";

import type { components, paths } from "./generated/schema";

export type { components, paths };

/** Typed OpenAPI client. Paths/schemas come from `src/api/generated/schema.ts`. */
export type ApiClient = ReturnType<typeof createApiClient>;

export type CreateApiClientOptions = {
  /** Shared secret for `Authorization: Bearer` when the API has API_AUTH_TOKEN set. */
  authToken?: string;
};

/**
 * Build a typed API client.
 *
 * Defaults to `NEXT_PUBLIC_API_BASE_URL`, falling back to the Prism mock on
 * `:4010` (see CONTRIBUTING.md). Pass an explicit base URL in tests.
 *
 * When the backend has `API_AUTH_TOKEN` configured, pass `authToken` (or set
 * `TEMFLOWRAL_API_TOKEN` in Node/test environments). Do not put secrets in
 * `NEXT_PUBLIC_*` — see SECURITY.md.
 */
export function createApiClient(
  baseUrl: string = getApiBaseUrl(),
  options: CreateApiClientOptions = {},
) {
  const token =
    options.authToken?.trim() ||
    process.env.TEMFLOWRAL_API_TOKEN?.trim() ||
    "";
  const headers: Record<string, string> = {};
  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }
  return createClient<paths>({ baseUrl, headers });
}
