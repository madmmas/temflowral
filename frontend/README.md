# temflowral frontend

Next.js (App Router) + TypeScript UI for authoring and running workflow graphs.

This app is a **reference implementation** of a temflowral canvas. It is not
published as an npm package and is not an embeddable hosted designer. If you
need a canvas in another product, integrate against the OpenAPI contract and
`GET /node-types` (see [`docs/adr/001-canvas-packaging.md`](../docs/adr/001-canvas-packaging.md)).

## Prerequisites

- Node 20+
- npm

## Setup

```bash
cd frontend
npm ci
cp .env.example .env.local   # optional; defaults to Prism on :4010
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

## Scripts

| Script | Purpose |
|---|---|
| `npm run dev` | Next.js development server |
| `npm run build` | Production build |
| `npm run start` | Serve production build |
| `npm run lint` | ESLint |
| `npm test` | Vitest (add `-- --run` for CI/Makefile) |
| `npm run generate` | Regenerate typed OpenAPI client from `../api/openapi.yaml` |
| `npm run e2e:install` | Install Playwright's Chromium browser once |
| `npm run e2e` | Run Playwright; automatically starts Prism + Next.js |
| `npm run e2e:ui` | Open Playwright's interactive test UI |

Point `NEXT_PUBLIC_API_BASE_URL` at the Prism mock (`http://127.0.0.1:4010`)
or the local backend (`http://127.0.0.1:8080`). See the repo
`CONTRIBUTING.md` for mock-server details.

If the backend has `API_AUTH_TOKEN` set, pass the token into
`createApiClient(url, { authToken })` or set `TEMFLOWRAL_API_TOKEN` for
Node/test processes. Never put the shared secret in `NEXT_PUBLIC_*` (see
`SECURITY.md`).

## Typed API client

Generated types live in `src/api/generated/` (do not hand-edit). Use the
wrapper instead of raw `fetch`:

```ts
import { createApiClient } from "@/api";

const api = createApiClient(); // defaults to NEXT_PUBLIC_API_BASE_URL / Prism
const { data, error } = await api.GET("/node-types");
```

From the repo root, `make generate` refreshes both the Go server interfaces and
this TypeScript client.

## Playwright

Install Chromium once, then run the isolated UI tests:

```bash
npm run e2e:install
npm run e2e
```

Playwright starts the pinned Prism mock on `http://127.0.0.1:4010` and Next.js
on `http://127.0.0.1:3000`, with the frontend pointed at Prism. Override either
server when needed:

```bash
API_BASE_URL=http://127.0.0.1:8080 \
PLAYWRIGHT_BASE_URL=http://127.0.0.1:3000 \
npm run e2e
```

When an override is set, Playwright assumes that server is already running.
Real-backend runs are opt-in; the default remains contract-backed Prism.
The happy-path spec only supplies Prism's missing terminal polling response;
when `API_BASE_URL` is set, all requests—including polling—use the real API.
