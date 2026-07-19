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

# Run golangci-lint in the given module directory through the shared pinned
# runner, keeping pre-commit, Makefile, and CI expectations aligned.
run_golangci_lint() {
  local dir="$1"
  local repo_root

  if [[ "${dir}" != "backend" ]]; then
    echo "lint: unsupported Go module directory: ${dir}" >&2
    return 1
  fi

  repo_root="$(git rev-parse --show-toplevel)"
  "${repo_root}/scripts/run-golangci-lint.sh"
}
