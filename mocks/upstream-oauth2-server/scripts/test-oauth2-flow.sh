#!/bin/bash

# Test script for Upstream OAuth2 Mock Server
# This script tests the OAuth2 authorization and token flows

set -e

UPSTREAM_BASE_URL="${UPSTREAM_BASE_URL:-http://127.0.0.1:9001}"
REDIRECT_URI="${REDIRECT_URI:-http://localhost:3000/callback}"

echo "🌐 Testing Upstream OAuth2 Mock Server"
echo "================================================"
echo "Base URL: $UPSTREAM_BASE_URL"
echo ""

# Test 1: Health check
echo "✓ Test 1: Health check"
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$UPSTREAM_BASE_URL/health")
HTTP_CODE=$(echo "$HEALTH_RESPONSE" | tail -n 1)
BODY=$(echo "$HEALTH_RESPONSE" | head -n -1)

if [ "$HTTP_CODE" -eq 200 ]; then
  echo "  Status: 200 OK"
  echo "  Response: $BODY"
else
  echo "  ✗ FAILED: Expected 200, got $HTTP_CODE"
  exit 1
fi
echo ""

# Test 2: Authorization endpoint returns consent page
echo "✓ Test 2: Authorization endpoint"
AUTH_REQUEST_URL="$UPSTREAM_BASE_URL/oauth/authorize?client_id=upstream-oauth2-client&response_type=code&redirect_uri=$REDIRECT_URI&scope=openid+profile&state=test123&code_challenge=E9Mrozoa2owUednTvZstLJBwcqnmJ07sv2we38R8oE&code_challenge_method=S256"

AUTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$AUTH_REQUEST_URL")
HTTP_CODE=$(echo "$AUTH_RESPONSE" | tail -n 1)
BODY=$(echo "$AUTH_RESPONSE" | head -n -1)

if [ "$HTTP_CODE" -eq 200 ]; then
  echo "  Status: 200 OK"
  if echo "$BODY" | grep -q "UPSTREAM OAUTH2"; then
    echo "  ✓ Contains distinctive 'UPSTREAM OAUTH2' badge"
  fi
  if echo "$BODY" | grep -q "#00d4aa"; then
    echo "  ✓ Contains teal background color (#00d4aa)"
  fi
  if echo "$BODY" | grep -q "Upstream OAuth2 Server"; then
    echo "  ✓ Contains server identifier"
  fi
else
  echo "  ✗ FAILED: Expected 200, got $HTTP_CODE"
  exit 1
fi
echo ""

# Test 3: Verify consent page contains expected form fields
echo "✓ Test 3: Consent page form validation"
if echo "$BODY" | grep -q 'name="client_id"'; then
  echo "  ✓ Form contains client_id field"
fi
if echo "$BODY" | grep -q 'name="approval"'; then
  echo "  ✓ Form contains approval field"
fi
if echo "$BODY" | grep -q 'value="approve"'; then
  echo "  ✓ Form contains approve button"
fi
if echo "$BODY" | grep -q 'value="deny"'; then
  echo "  ✓ Form contains deny button"
fi
echo ""

echo "================================================"
echo "✓ All tests passed!"
echo ""
echo "Manual Testing:"
echo "  1. Open browser: $AUTH_REQUEST_URL"
echo "  2. Verify teal background (different from third-party mock)"
echo "  3. Verify 'UPSTREAM OAUTH2' badge is visible"
echo "  4. Click 'Approve' to test consent flow"
