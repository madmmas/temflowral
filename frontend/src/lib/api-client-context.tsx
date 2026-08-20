"use client";

import {
  createContext,
  useCallback,
  useContext,
  useMemo,
  type ReactNode,
} from "react";

import { createApiClient, type ApiClient, type CreateApiClientOptions } from "@/api";

export type WorkflowBuilderApiConfig = {
  /** OpenAPI base URL (defaults to NEXT_PUBLIC_API_BASE_URL / Prism). */
  apiBaseUrl?: string;
  /** Static Bearer token when the API has API_AUTH_TOKEN set. */
  authToken?: string;
  /** Prefer over authToken when the host refreshes credentials. */
  getAccessToken?: () => string | Promise<string | undefined>;
};

type ApiClientContextValue = {
  config: WorkflowBuilderApiConfig;
  createClient: () => ApiClient;
};

const ApiClientContext = createContext<ApiClientContextValue | null>(null);

type ApiClientProviderProps = WorkflowBuilderApiConfig & {
  children: ReactNode;
};

/**
 * Supplies API base URL / auth to the workflow builder and its hooks (#sidecar).
 */
export function ApiClientProvider({
  apiBaseUrl,
  authToken,
  getAccessToken,
  children,
}: ApiClientProviderProps) {
  const config = useMemo(
    () => ({ apiBaseUrl, authToken, getAccessToken }),
    [apiBaseUrl, authToken, getAccessToken],
  );

  const createClient = useCallback(() => {
    const options: CreateApiClientOptions = {
      authToken,
      getAccessToken,
    };
    return createApiClient(apiBaseUrl, options);
  }, [apiBaseUrl, authToken, getAccessToken]);

  const value = useMemo(
    () => ({ config, createClient }),
    [config, createClient],
  );

  return (
    <ApiClientContext.Provider value={value}>{children}</ApiClientContext.Provider>
  );
}

/** Prefer the nearest WorkflowBuilder provider; fall back to env defaults. */
export function useCreateApiClient(): () => ApiClient {
  const ctx = useContext(ApiClientContext);
  return ctx?.createClient ?? (() => createApiClient());
}
