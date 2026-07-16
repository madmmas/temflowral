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
