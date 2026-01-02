#!/bin/bash
# Register mock OAuth2 service with identity broker admin API

set -e

ADMIN_API="http://localhost:14000/api"
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "========================================="
echo "Registering Mock OAuth2 Service"
echo "========================================="

# Check broker admin API
echo -n "Checking admin API... "
if curl -s -f "http://localhost:14000/health" > /dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    echo "Error: Start broker first with: just run"
    exit 1
fi

# Check mock server
echo -n "Checking mock OAuth2 server... "
if curl -s -f "http://localhost:9000/health" > /dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    echo "Error: Start mock server first with: just mock-third-party-oauth2-start"
    exit 1
fi

# Register service
echo "Registering mock OAuth2 service..."
SERVICE_RESPONSE=$(curl -s -X POST "${ADMIN_API}/services" \
  -H "Content-Type: application/json" \
  -H "X-Remote-User: admin@example.com" \
  -w "\n%{http_code}" \
  -d '{
    "display_name": "Mock OAuth2 Service (Dev)",
    "client_id": "mock-oauth2-client-dev",
    "client_secret": "mock-oauth2-secret-12345",
    "issuer_uri": "http://localhost:9000",
    "discovery": {"enable_discovery": false},
    "endpoints": {
      "token_endpoint": "http://localhost:9000/oauth/token",
      "authorize_endpoint": "http://localhost:9000/oauth/authorize"
    },
    "scopes": [
      {"scope_value": "profile", "description": "Access user profile"},
      {"scope_value": "email", "description": "Access email address"},
      {"scope_value": "read", "description": "Read data"},
      {"scope_value": "write", "description": "Write data"}
    ]
  }')

HTTP_CODE=$(echo "$SERVICE_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$SERVICE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Mock OAuth2 service registered${NC}"
    SERVICE_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "  Service ID: $SERVICE_ID"
    echo ""
    echo "Next steps:"
    echo "  1. Open: http://localhost:3000/api/third-party/$SERVICE_ID/oauth2/authorize?redirect_uri=http%3A%2F%2Flocalhost%3A3000%2Fconsent"
    echo "  2. Approve consent on mock OAuth2 page"
    echo "  3. Session established!"
else
    echo -e "${RED}✗ Registration failed${NC}"
    echo "  HTTP Status: $HTTP_CODE"
    echo "  Response: $RESPONSE_BODY"
    exit 1
fi
