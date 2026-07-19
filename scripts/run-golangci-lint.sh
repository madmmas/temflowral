#!/usr/bin/env bash
# Run the same golangci-lint version used by CI, without requiring a separate
# local installation. A matching installed binary is preferred for offline use.
set -euo pipefail

GOLANGCI_LINT_VERSION="v2.12.2"
EXPECTED_VERSION="${GOLANGCI_LINT_VERSION#v}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

cd "${REPO_ROOT}/backend"

if command -v golangci-lint >/dev/null 2>&1; then
  version_output="$(golangci-lint version 2>/dev/null || true)"
  if [[ "${version_output}" == *"version ${EXPECTED_VERSION}"* ]]; then
    exec golangci-lint run "$@"
  fi
  echo "lint: installed golangci-lint is not ${EXPECTED_VERSION}; using pinned go run." >&2
fi

if ! command -v go >/dev/null 2>&1; then
  echo "lint: Go is required when golangci-lint ${EXPECTED_VERSION} is not installed." >&2
  exit 1
fi

echo "lint: using golangci-lint ${GOLANGCI_LINT_VERSION} via go run." >&2
exec go run \
  "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}" \
  run "$@"
