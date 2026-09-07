#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

OPA_IMAGE="openpolicyagent/opa:latest"
BUNDLE_MOUNT="-v $(pwd)/bundle:/bundle:ro"

echo "Checking..."
docker run --rm $BUNDLE_MOUNT $OPA_IMAGE check --bundle /bundle

mkdir -p "$(pwd)/bundles"

echo "Building..."
docker run --rm \
  -w /bundle \
  -v "$(pwd)/bundle:/bundle:ro" \
  -v "$(pwd)/bundles:/output" \
  $OPA_IMAGE build -o /output/mcp-authz.tar.gz .

echo "Done."