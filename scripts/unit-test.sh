#!/usr/bin/env bash
# Run unit tests with coverage for pkg and each backend module.
# Generates coverage files under ./coverage/
set -euo pipefail

COVER_DIR=$(mktemp -d -t coverage-XXXXXX)
PKG_COVER="$COVER_DIR/coverage-pkg.out"
ALL_COVER="$COVER_DIR/coverage-all.out"

mkdir -p "$COVER_DIR"

if ! command -v go >/dev/null 2>&1; then
  echo "go not found in PATH. Install Go and retry."
  exit 1
fi

echo "==> Running unit tests for core packages (./pkg/...)"
echo "Output coverage: $PKG_COVER"
go test ./pkg/... -covermode=atomic -coverpkg=./... -coverprofile="$PKG_COVER"

# Collect backend coverage files
backend_cov_files=()

echo
echo "==> Running backend tests (each backend/* with go.mod)"
for d in backend/*/; do
  if [ -f "${d}go.mod" ]; then
    bkname=$(basename "$d")
    out="$COVER_DIR/coverage-backend-${bkname}.out"
    echo "-> Testing backend: $bkname (coverage -> $out)"
    # Create directory for coverprofile
    mkdir -p "$COVER_DIR"

    # Use a subshell; keep going on failure but report it.
    if ! (cd "$d" && go test ./... -covermode=atomic -coverprofile="$out" -timeout=300s); then
      echo "!! Tests failed for backend: $bkname (see output above). Continuing with next backend."
      # still record the (possibly partial) coverage file if it exists
    fi
    # add to list if file exists
    if [ -f "$out" ]; then
      backend_cov_files+=("$out")
    fi
  fi
done

# Merge coverage files if gocovmerge is available, otherwise concatenate (not recommended but ok)
echo
if [ "${#backend_cov_files[@]}" -gt 0 ]; then
  echo "Merging coverage files with local merge tool -> $ALL_COVER"
  # use the included Go merger; it requires Go in PATH
  go run ./scripts/merge_coverage.go -o "$ALL_COVER" "$PKG_COVER" "${backend_cov_files[@]}" || {
    echo "Merging failed, falling back to simple concat -> $ALL_COVER"
    awk 'FNR==1 && NR!=1{next} {print}' "$PKG_COVER" "${backend_cov_files[@]}" > "$ALL_COVER"
  }
else
  echo "No backend coverage files found. Copying pkg coverage to $ALL_COVER"
  cp "$PKG_COVER" "$ALL_COVER"
fi

# Generate HTML reports if go tool cover available
  echo "Generating HTML reports..."
  go tool cover -html="$PKG_COVER" -o "$COVER_DIR/coverage-pkg.html" || true
  go tool cover -html="$ALL_COVER" -o "$COVER_DIR/coverage-all.html" || true
  echo "HTML reports: $COVER_DIR/coverage-pkg.html , $COVER_DIR/coverage-all.html"

echo
echo "Coverage outputs:"
ls -1 "$COVER_DIR" || true
echo
echo "Done."