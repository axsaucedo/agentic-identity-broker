#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

OPA_IMAGE="${OPA_IMAGE:-openpolicyagent/opa:latest}"
BUNDLE_DIR="$(pwd)/bundle"
OUTPUT_DIR="$(pwd)/bundles"

echo "Checking..."
docker run --rm -v "${BUNDLE_DIR}:/bundle:ro" "$OPA_IMAGE" check --bundle /bundle

mkdir -p "$OUTPUT_DIR"

echo "Building..."
docker run --rm \
  -w /bundle \
  -v "${BUNDLE_DIR}:/bundle:ro" \
  -v "${OUTPUT_DIR}:/output" \
  "$OPA_IMAGE" build -o /output/mcp-authz.tar.gz .

echo "Done."
