# Embed the workflow builder via Module Federation

Amends [ADR-001](adr/001-canvas-packaging.md) — see
[ADR-002](adr/002-module-federation-canvas.md).

The engine still runs as an HTTP sidecar ([sidecar.md](sidecar.md)). This
guide covers loading the **canvas remote** into a host SPA.

## What is exposed

| Expose | Module | Export |
| --- | --- | --- |
| `temflowralCanvas/WorkflowBuilder` | `./WorkflowBuilder` | `WorkflowBuilder` component + `WorkflowBuilderProps` |

### Props contract

```ts
type WorkflowBuilderProps = {
  apiBaseUrl?: string;
  authToken?: string;
  getAccessToken?: () => string | Promise<string | undefined>;
  initialGraphId?: string | null;
  /** Default true in the reference app; set false when embedded. */
  syncGraphQueryParam?: boolean;
  onGraphSaved?: (graphId: string) => void;
  onRunCompleted?: (run: Run) => void;
  className?: string;
};
```

Always pass `apiBaseUrl` to your sidecar API. Prefer `getAccessToken` over
baking secrets into the bundle. Set `API_CORS_ORIGINS` on the sidecar to your
host origin.

## Shared dependencies (singletons)

Align versions with the remote (see `packages/canvas-remote/package.json`):

- `react` / `react-dom` `19.1.x`
- `@xyflow/react` `12.8.x`

Mark them `singleton: true` in the host MF config or the canvas will break.

## Build and serve the remote

```sh
make canvas-remote-build
# artifacts in packages/canvas-remote/dist/remoteEntry.js
```

Dev server (CORS open):

```sh
cd packages/canvas-remote && npm install && npm run dev
# http://127.0.0.1:3002/remoteEntry.js
```

Docker static image (optional):

```sh
docker build -f packages/canvas-remote/Dockerfile -t temflowral-canvas-remote .
```

## Host configuration (Rspack / webpack)

```js
new ModuleFederationPlugin({
  name: "hostApp",
  remotes: {
    temflowralCanvas:
      "temflowralCanvas@http://127.0.0.1:3002/remoteEntry.js",
  },
  shared: {
    react: { singleton: true, requiredVersion: "19.1.0" },
    "react-dom": { singleton: true, requiredVersion: "19.1.0" },
    "@xyflow/react": { singleton: true, requiredVersion: "12.8.6" },
  },
});
```

```tsx
import { lazy, Suspense } from "react";

const WorkflowBuilder = lazy(() =>
  import("temflowralCanvas/WorkflowBuilder").then((m) => ({
    default: m.WorkflowBuilder,
  })),
);

export function WorkflowsPage() {
  return (
    <Suspense fallback={<p>Loading builder…</p>}>
      <WorkflowBuilder
        apiBaseUrl="https://temflowral.internal"
        getAccessToken={() => fetchToken()}
        syncGraphQueryParam={false}
        className="h-[80vh]"
      />
    </Suspense>
  );
}
```

## Next.js hosts

Use **webpack** (or Rspack) with `@module-federation/nextjs-mf` /
`@module-federation/enhanced`. **Turbopack** is not a supported MF host path
yet — run `next dev --webpack` / production webpack builds for federated pages.

The reference app in this repo imports `WorkflowBuilder` **locally** so
Playwright e2e stays self-contained. Prefer the remote URL in real host apps.

## CSS / Tailwind

The remote injects its styles (including Tailwind + React Flow) via
`style-loader`. Host global CSS may still conflict; scope host layout outside
the builder root and prefer the `className` prop for height.

## Versioning

Ship `remoteEntry.js` (+ chunks) from a versioned URL or image tag
(`temflowral-canvas-remote:0.1.0`). Pin the remote URL in the host; bump
deliberately when props or shared deps change.
