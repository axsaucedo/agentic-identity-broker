#!/usr/bin/env bash
# Helm schema validation tests for values.schema.json
# Tests that valid configurations pass and invalid configurations are rejected.
# Usage: ./scripts/helm-schema-test.sh
set -euo pipefail

CHART="charts/agentic-identity-broker"
PASS=0
FAIL=0

pass() { echo "  ✓ $1"; PASS=$((PASS+1)); }
fail() { echo "  ✗ $1"; FAIL=$((FAIL+1)); }

# Run helm lint; return 0 if it passes, 1 if it fails.
lint_ok()  { helm lint "$CHART" "$@" >/dev/null 2>&1; }
lint_err() { ! helm lint "$CHART" "$@" >/dev/null 2>&1; }

echo "==> Helm schema validation tests"
echo ""

echo "--- Valid configurations (should pass) ---"

if lint_ok -f "$CHART/ci/local-values.yaml"; then
  pass "local mode without proxy fields"
else
  fail "local mode without proxy fields"
fi

if lint_ok -f "$CHART/ci/proxy-values.yaml"; then
  pass "proxy mode with all three upstream fields"
else
  fail "proxy mode with all three upstream fields"
fi

if lint_ok -f "$CHART/ci/hybrid-values.yaml"; then
  pass "hybrid mode with all three upstream fields"
else
  fail "hybrid mode with all three upstream fields"
fi

echo ""
echo "--- Invalid configurations (should fail) ---"

if lint_err -f "$CHART/ci/proxy-values.yaml" \
    --set broker.oauth2AuthorizationServer.proxy.upstreamIssuerUri=""; then
  pass "proxy mode with missing upstreamIssuerUri rejected"
else
  fail "proxy mode with missing upstreamIssuerUri rejected"
fi

if lint_err -f "$CHART/ci/proxy-values.yaml" \
    --set broker.oauth2AuthorizationServer.proxy.upstreamAuthorizeEndpoint=""; then
  pass "proxy mode with missing upstreamAuthorizeEndpoint rejected"
else
  fail "proxy mode with missing upstreamAuthorizeEndpoint rejected"
fi

if lint_err -f "$CHART/ci/proxy-values.yaml" \
    --set broker.oauth2AuthorizationServer.proxy.upstreamTokenEndpoint=""; then
  pass "proxy mode with missing upstreamTokenEndpoint rejected"
else
  fail "proxy mode with missing upstreamTokenEndpoint rejected"
fi

if lint_err -f "$CHART/ci/hybrid-values.yaml" \
    --set broker.oauth2AuthorizationServer.proxy.upstreamIssuerUri=""; then
  pass "hybrid mode with missing upstreamIssuerUri rejected"
else
  fail "hybrid mode with missing upstreamIssuerUri rejected"
fi

# mode omitted from a partial override: Helm merges with values.yaml, so the
# default mode fills in — schema cannot catch this. Test the empty-string case
# instead, which Helm does NOT merge away and the schema rejects explicitly.
if lint_err --set broker.oauth2AuthorizationServer.mode=""; then
  pass "empty mode string rejected"
else
  fail "empty mode string rejected"
fi

echo ""
echo "==> Results: $PASS passed, $FAIL failed"
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
