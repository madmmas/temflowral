# Security policy

## Supported versions

| Version | Supported |
|---------|-----------|
| 0.x.x   | Yes       |

## Reporting a vulnerability

**Do not open a public GitHub issue for security vulnerabilities.**

Please use GitHub's private vulnerability reporting:
1. Go to the Security tab of this repository
2. Click "Report a vulnerability"
3. Fill in the details

I aim to respond within 48 hours and will keep you updated
on the fix timeline.

## Scope

temflowral is a demonstration project. The primary security
concerns are:
- Arbitrary HTTP requests made by the HTTP activity node
- Template injection in node configuration fields
- Server-side request forgery via user-supplied URLs
- Unauthenticated or overly trusted access to the HTTP API

## Trust boundary (no tenant isolation)

temflowral is a **single-trust-domain engine**. It does **not** enforce
multi-tenant isolation, ownership, or per-caller ACLs on graphs, runs, or
signals.

- Knowing a graph ID or run ID is sufficient to fetch, run, or signal that
  resource once a caller can reach the API with a valid shared secret (when
  auth is enabled).
- Callers (API gateways, BFF layers, product backends) **must** authorize
  which principals may access which graph/run/signal IDs **before** forwarding
  traffic to temflowral.
- Do not assume UUID opacity is an access-control mechanism.

## API authentication baseline

Production and shared deployments should enable service-to-service auth:

```sh
API_AUTH_TOKEN='long-random-secret'
```

When `API_AUTH_TOKEN` is set, every OpenAPI route requires:

```http
Authorization: Bearer <API_AUTH_TOKEN>
```

Missing or incorrect tokens return `401` with `code: unauthorized`.
`GET /docs` and `GET /openapi.yaml` stay reachable without a token so operators
can read the contract.

When `API_AUTH_TOKEN` is unset, the API runs in **open mode** (local
development, Prism mocks, and compose defaults). Do not expose an open-mode
server on an untrusted network.

### mTLS

Mutual TLS is supported as a **deployment** option: terminate client
certificates at a reverse proxy or service mesh (nginx, Caddy, Envoy, ingress)
in front of the Go process. temflowral does not implement application-level
mTLS identity checks in v0.x; combine proxy mTLS with `API_AUTH_TOKEN` when you
want defense in depth.

Do not put shared secrets in `NEXT_PUBLIC_*` frontend env vars — the reference
canvas is not a place to embed service credentials. Service callers and BFFs
should attach the Bearer header server-side.

## Interpreter / upgrade compatibility

Upgrading the temflowral backend (graph interpreter + worker) can change
execution semantics for graphs that already exist. Watch for:

- **Node-type registry** — new required config fields, removed types, or
  stricter validation on `POST /graphs` / `POST .../run` (`GET /node-types`
  is the discovery surface).
- **HTTP allowlist / SSRF policy** — `HTTP_ALLOWED_HOSTS` and destination
  checks may reject URLs that previously worked.
- **Templating** — path resolution and forbidden contexts (e.g. wait config)
  can tighten.
- **Signals** — wait signal names and `POST /runs/{id}/signal` matching rules.
- **In-flight Temporal workflows** — workflow/activity code changes can be
  non-replay-safe for runs started on an older binary; drain or pin worker
  versions before rolling breaking interpreter changes.

Pin versions in production, re-validate stored graphs after upgrades, and treat
Temporal worker rollouts like any other durable-execution code change.

## HTTP activity controls

HTTP activities deny all destinations by default. Operators must set
`HTTP_ALLOWED_HOSTS` to a comma-separated list of exact hostnames (no schemes,
ports, or wildcards), for example:

```sh
HTTP_ALLOWED_HOSTS=api.example.com,hooks.example.com
```

The worker accepts only HTTP(S), disables environment proxies, revalidates
redirects, resolves and validates the address used for the actual connection,
and rejects loopback, private, link-local, multicast, and unspecified
addresses. Request and response bodies are limited to 1 MiB, response headers
to 64 KiB, and requests to 20 seconds. Hop-by-hop/transport-controlled request
headers are rejected.

Template interpolation is supported for node config string leaves using the
minimal syntax `{{ nodes.<nodeId>.output.<path> }}`. Paths are resolved at
execution time from active predecessor outputs only (no env, filters, or
expression language). Templates are not allowed in wait node config. For HTTP
nodes, the fully rendered URL, headers, and body are always revalidated through
the same allowlist, SSRF, header, and size policy as concrete values — never
treat a templated string as trusted without that second pass. Activity retries
remain disabled so Temporal does not automatically repeat POST/PATCH or other
side-effecting requests.

If you find a vulnerability in a dependency, please report it
to that project directly. You can also open a Dependabot alert
via the Security tab.
