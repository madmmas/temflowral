# temflowral frontend

Next.js (App Router) + TypeScript UI for authoring and running workflow graphs.

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

Point `NEXT_PUBLIC_API_BASE_URL` at the Prism mock (`http://127.0.0.1:4010`)
or the local backend (`http://127.0.0.1:8080`). See the repo
`CONTRIBUTING.md` for mock-server details.
