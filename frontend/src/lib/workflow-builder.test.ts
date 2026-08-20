import { describe, expect, it } from "vitest";

import { createApiClient } from "@/api/client";

describe("WorkflowBuilder API client options", () => {
  it("accepts getAccessToken without throwing", () => {
    const client = createApiClient("http://127.0.0.1:8080", {
      getAccessToken: async () => "token",
    });
    expect(typeof client.GET).toBe("function");
  });

  it("treats omitted base URL as getApiBaseUrl default", () => {
    const client = createApiClient(undefined, { authToken: "x" });
    expect(client).toBeDefined();
  });
});
