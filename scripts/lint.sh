#!/usr/bin/env bash
# Small script to run golangci-lint --fix at repo root and in each backend/* module
set -euo pipefail

if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "golangci-lint not found. Install it (see .github/copilot-instructions.md) and retry."
  exit 1
fi

echo "Running golangci-lint --fix at repo root..."
golangci-lint run --fix ./...

for d in backend/*/; do
  if [ -f "${d}go.mod" ]; then
    echo "Running golangci-lint --fix in ${d}..."
    (cd "$d" && golangci-lint run --fix ./...) || echo "golangci-lint failed in ${d} (continue)"
  fi
done

echo "golangci-lint --fix completed."