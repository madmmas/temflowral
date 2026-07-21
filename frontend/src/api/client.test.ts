import { afterEach, describe, expect, it } from "vitest";

import { createApiClient } from "./client";
import { getApiBaseUrl } from "@/lib/api-base-url";

describe("createApiClient", () => {
  const original = process.env.NEXT_PUBLIC_API_BASE_URL;

  afterEach(() => {
    if (original === undefined) {
      delete process.env.NEXT_PUBLIC_API_BASE_URL;
    } else {
      process.env.NEXT_PUBLIC_API_BASE_URL = original;
    }
  });

  it("exposes typed GET and POST helpers", () => {
    const client = createApiClient("http://127.0.0.1:4010");
    expect(typeof client.GET).toBe("function");
    expect(typeof client.POST).toBe("function");
  });

  it("defaults base URL via getApiBaseUrl (Prism mock)", () => {
    delete process.env.NEXT_PUBLIC_API_BASE_URL;
    expect(getApiBaseUrl()).toBe("http://127.0.0.1:4010");
    const client = createApiClient();
    expect(client).toBeDefined();
  });

  it("accepts an explicit base URL override", () => {
    const client = createApiClient("http://127.0.0.1:8080");
    expect(client).toBeDefined();
  });

  it("accepts an authToken option without throwing", () => {
    const client = createApiClient("http://127.0.0.1:8080", {
      authToken: "test-secret",
    });
    expect(client).toBeDefined();
  });
});
