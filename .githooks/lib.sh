#!/usr/bin/env bash
# Shared helpers for versioned git hooks.

# Space-separated list of branches that must only move via pull request.
PROTECTED_BRANCHES="main master"

current_branch() {
  git branch --show-current 2>/dev/null || true
}

is_protected_branch() {
  local branch="$1"
  local protected
  for protected in ${PROTECTED_BRANCHES}; do
    if [[ "${branch}" == "${protected}" ]]; then
      return 0
    fi
  done
  return 1
}

protected_branch_from_ref() {
  local ref="$1"
  ref="${ref#refs/heads/}"
  printf '%s' "${ref}"
}

# Pinned to match the version used by the go-lint CI job.
GOLANGCI_LINT_VERSION="v2.12.2"

# Run golangci-lint in the given module directory. Prefers a matching
# golangci-lint on PATH; otherwise falls back to `go run` at the pinned
# version so local checks mirror CI without a separate install step.
run_golangci_lint() {
  local dir="$1"

  if command -v golangci-lint >/dev/null 2>&1; then
    (cd "${dir}" && golangci-lint run)
    return $?
  fi

  if command -v go >/dev/null 2>&1; then
    echo "pre-commit: golangci-lint not found on PATH; using 'go run ${GOLANGCI_LINT_VERSION}'." >&2
    (cd "${dir}" && go run "github.com/golangci/golangci-lint/v2/cmd/golangci-lint@${GOLANGCI_LINT_VERSION}" run)
    return $?
  fi

  echo "pre-commit: neither golangci-lint nor go is installed; cannot lint ${dir}." >&2
  return 1
}
