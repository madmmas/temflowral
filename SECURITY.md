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

Template interpolation is intentionally not implemented. Activity retries are
disabled so Temporal does not automatically repeat POST/PATCH or other
side-effecting requests. Never add URL/header/body templating without
revalidating the fully rendered request through the same URL and header policy.

If you find a vulnerability in a dependency, please report it
to that project directly. You can also open a Dependabot alert
via the Security tab.
