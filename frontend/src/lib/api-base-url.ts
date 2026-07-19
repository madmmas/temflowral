/** Default matches the Prism mock documented in CONTRIBUTING.md. */
const DEFAULT_API_BASE_URL = "http://127.0.0.1:4010";

export function getApiBaseUrl(): string {
  return process.env.NEXT_PUBLIC_API_BASE_URL ?? DEFAULT_API_BASE_URL;
}
