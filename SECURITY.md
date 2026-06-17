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

If you find a vulnerability in a dependency, please report it
to that project directly. You can also open a Dependabot alert
via the Security tab.