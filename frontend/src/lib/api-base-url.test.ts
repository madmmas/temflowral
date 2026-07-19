import { afterEach, describe, expect, it } from "vitest";

import { getApiBaseUrl } from "./api-base-url";

describe("getApiBaseUrl", () => {
  const original = process.env.NEXT_PUBLIC_API_BASE_URL;

  afterEach(() => {
    if (original === undefined) {
      delete process.env.NEXT_PUBLIC_API_BASE_URL;
    } else {
      process.env.NEXT_PUBLIC_API_BASE_URL = original;
    }
  });

  it("defaults to the Prism mock URL", () => {
    delete process.env.NEXT_PUBLIC_API_BASE_URL;
    expect(getApiBaseUrl()).toBe("http://127.0.0.1:4010");
  });

  it("reads NEXT_PUBLIC_API_BASE_URL when set", () => {
    process.env.NEXT_PUBLIC_API_BASE_URL = "http://127.0.0.1:8080";
    expect(getApiBaseUrl()).toBe("http://127.0.0.1:8080");
  });
});
