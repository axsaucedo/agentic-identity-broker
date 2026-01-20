#!/bin/bash
# Combined setup script: Register mock OAuth2 service AND seed sample data
# This ensures the mock service is available before creating agents that require it

set -e

# Allow override via environment variables for container support
ADMIN_API="${ADMIN_API:-http://localhost:14000/api}"
BROKER_HEALTH_URL="${BROKER_HEALTH_URL:-http://localhost:14000/health}"
# Separate token and authorize endpoints to support service names in Docker Compose
MOCK_SERVER_TOKEN_URL="${MOCK_SERVER_TOKEN_URL:-http://localhost:9000}"
MOCK_SERVER_AUTHORIZE_URL="${MOCK_SERVER_AUTHORIZE_URL:-http://localhost:9000}"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Setting Up Development Environment"
echo "========================================="
echo ""

# Check if admin API is accessible
echo -n "Checking admin API connectivity... "
if curl -s -f "${BROKER_HEALTH_URL}" > /dev/null; then
    echo -e "${GREEN}OK${NC}"
else
    echo -e "${RED}FAILED${NC}"
    echo "Error: Admin API is not accessible at ${ADMIN_API}"
    echo "Please ensure the identity broker is running with admin API on port 14000"
    exit 1
fi
echo ""

# ==========================================
# Step 1: Register Mock OAuth2 Service
# ==========================================
echo "========================================="
echo "Step 1: Registering Mock OAuth2 Service"
echo "========================================="
echo ""

echo "Registering mock OAuth2 service..."
SERVICE_RESPONSE=$(curl -s -X POST "${ADMIN_API}/services" \
  -H "Content-Type: application/json" \
  -H "X-Remote-User: admin@example.com" \
  -w "\n%{http_code}" \
  -d "{
    \"display_name\": \"Mock OAuth2 Service (Dev)\",
    \"client_id\": \"mock-oauth2-client-dev\",
    \"client_secret\": \"mock-oauth2-secret-12345\",
    \"issuer_uri\": \"http://localhost:9000\",
    \"discovery\": {\"enable_discovery\": false},
    \"endpoints\": {
      \"token_endpoint\": \"${MOCK_SERVER_TOKEN_URL}/oauth/token\",
      \"authorize_endpoint\": \"${MOCK_SERVER_AUTHORIZE_URL}/oauth/authorize\"
    },
    \"scopes\": [
      {\"scope_value\": \"profile\", \"description\": \"Access user profile\"},
      {\"scope_value\": \"email\", \"description\": \"Access email address\"},
      {\"scope_value\": \"read\", \"description\": \"Read data\"},
      {\"scope_value\": \"write\", \"description\": \"Write data\"}
    ]
  }")

HTTP_CODE=$(echo "$SERVICE_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$SERVICE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Mock OAuth2 service registered${NC}"
    MOCK_SERVICE_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "  Service ID: $MOCK_SERVICE_ID"
elif [ "$HTTP_CODE" = "409" ]; then
    echo -e "${YELLOW}✓ Mock OAuth2 service already exists${NC}"
    # Query for existing service
    SERVICES=$(curl -s "${ADMIN_API}/services" -H "X-Remote-User: admin@example.com")
    MOCK_SERVICE_ID=$(echo "$SERVICES" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "  Service ID: $MOCK_SERVICE_ID"
else
    echo -e "${RED}✗ Failed to register Mock OAuth2 service${NC}"
    echo "  HTTP Status: $HTTP_CODE"
    echo "  Response: $RESPONSE_BODY"
    exit 1
fi
echo ""

# ==========================================
# Step 2: Seed Sample Data
# ==========================================
echo "========================================="
echo "Step 2: Seeding Sample Data"
echo "========================================="
echo ""

# Create Weather Assistant Agent
echo "Creating Weather Assistant agent..."
AGENT_RESPONSE=$(curl -s -X POST "${ADMIN_API}/agents" \
  -H "Content-Type: application/json" \
  -w "\n%{http_code}" \
  -d '{
    "client_id": "weather-assistant-001",
    "display_name": "Weather Assistant",
    "description": "AI assistant that helps you check weather forecasts and conditions",
    "governance_url": "https://example.com/weather-assistant/governance",
    "user_documentation_url": "https://example.com/weather-assistant/docs",
    "agent_interface_url": "https://example.com/weather-assistant"
  }')

HTTP_CODE=$(echo "$AGENT_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$AGENT_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Weather Assistant created${NC}"
    AGENT_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "  Agent ID: $AGENT_ID"
elif [ "$HTTP_CODE" = "409" ]; then
    echo -e "${YELLOW}✓ Weather Assistant already exists${NC}"
else
    echo -e "${RED}✗ Failed to create Weather Assistant${NC}"
    echo "  HTTP Status: $HTTP_CODE"
    exit 1
fi
echo ""

# Create Task Manager Agent
echo "Creating Task Manager agent..."
AGENT_RESPONSE=$(curl -s -X POST "${ADMIN_API}/agents" \
  -H "Content-Type: application/json" \
  -w "\n%{http_code}" \
  -d '{
    "client_id": "task-manager-agent",
    "display_name": "Task Manager Pro",
    "description": "Intelligent task management assistant that helps organize your work",
    "governance_url": "https://example.com/task-manager/governance",
    "user_documentation_url": "https://example.com/task-manager/docs"
  }')

HTTP_CODE=$(echo "$AGENT_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$AGENT_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Task Manager Pro created${NC}"
    AGENT_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "  Agent ID: $AGENT_ID"
elif [ "$HTTP_CODE" = "409" ]; then
    echo -e "${YELLOW}✓ Task Manager Pro already exists${NC}"
else
    echo -e "${RED}✗ Failed to create Task Manager Pro${NC}"
    echo "  HTTP Status: $HTTP_CODE"
    exit 1
fi
echo ""

# Create Weather API Service (before OAuth2 Test Client so we can reference it)
echo "Creating Weather API service..."
SERVICE_RESPONSE=$(curl -s -X POST "${ADMIN_API}/services" \
  -H "Content-Type: application/json" \
  -w "\n%{http_code}" \
  -d '{
    "display_name": "Weather API",
    "client_id": "weather-api-client",
    "client_secret": "weather-secret-12345",
    "issuer_uri": "https://auth.weather-api.example.com",
    "discovery": {
      "enable_discovery": false
    },
    "endpoints": {
      "token_endpoint": "https://auth.weather-api.example.com/oauth/token",
      "authorize_endpoint": "https://auth.weather-api.example.com/oauth/authorize"
    },
    "scopes": [
      {
        "scope_value": "weather.read",
        "description": "Read current weather conditions"
      },
      {
        "scope_value": "weather.forecast",
        "description": "Access weather forecasts"
      },
      {
        "scope_value": "weather.alerts",
        "description": "Receive weather alerts"
      }
    ]
  }')

HTTP_CODE=$(echo "$SERVICE_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$SERVICE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Weather API service created${NC}"
    WEATHER_SERVICE_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "  Service ID: $WEATHER_SERVICE_ID"
elif [ "$HTTP_CODE" = "409" ]; then
    echo -e "${YELLOW}✓ Weather API service already exists${NC}"
    # Query for existing service
    SERVICES=$(curl -s "${ADMIN_API}/services" -H "X-Remote-User: admin@example.com")
    WEATHER_SERVICE_ID=$(echo "$SERVICES" | grep -o '"id":"[^"]*"' | tail -1 | cut -d'"' -f4)
else
    echo -e "${RED}✗ Failed to create Weather API service${NC}"
    echo "  HTTP Status: $HTTP_CODE"
    exit 1
fi
echo ""

# Create OAuth2 Test Client Agent with Mock Service (mandatory) and Weather API (optional)
echo "Creating OAuth2 Test Client agent..."
AGENT_RESPONSE=$(curl -s -X POST "${ADMIN_API}/agents" \
  -H "Content-Type: application/json" \
  -w "\n%{http_code}" \
  -d "{
    \"client_id\": \"upstream-oauth2-client\",
    \"display_name\": \"Sample Agent\",
    \"description\": \"An agent that uses a thirdparty service\",
    \"governance_url\": \"https://example.com/oauth2-test/governance\",
    \"user_documentation_url\": \"https://example.com/oauth2-test/docs\",
    \"agent_interface_url\": \"http://localhost:3000\",
    \"service_requirements\": [
      {
        \"service_id\": \"${MOCK_SERVICE_ID}\",
        \"requirement_type\": \"mandatory\",
        \"required_scopes\": [\"profile\", \"email\"]
      },
      {
        \"service_id\": \"${WEATHER_SERVICE_ID}\",
        \"requirement_type\": \"optional\",
        \"required_scopes\": [\"weather.read\"]
      }
    ]
  }")

HTTP_CODE=$(echo "$AGENT_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$AGENT_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ OAuth2 Test Client created${NC}"
    AGENT_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "  Agent ID: $AGENT_ID"
    echo "  Service requirements:"
    echo "    - Mock OAuth2 Service (mandatory, scopes: profile, email)"
    echo "    - Weather API (optional, scopes: weather.read)"
elif [ "$HTTP_CODE" = "409" ]; then
    echo -e "${YELLOW}✓ OAuth2 Test Client already exists${NC}"
else
    echo -e "${RED}✗ Failed to create OAuth2 Test Client${NC}"
    echo "  HTTP Status: $HTTP_CODE"
    exit 1
fi
echo ""

# Create Calendar Service
echo "Creating Calendar API service..."
SERVICE_RESPONSE=$(curl -s -X POST "${ADMIN_API}/services" \
  -H "Content-Type: application/json" \
  -w "\n%{http_code}" \
  -d '{
    "display_name": "Calendar API",
    "client_id": "calendar-api-client",
    "client_secret": "calendar-secret-67890",
    "issuer_uri": "https://auth.calendar-api.example.com",
    "discovery": {
      "enable_discovery": false
    },
    "endpoints": {
      "token_endpoint": "https://auth.calendar-api.example.com/oauth/token",
      "authorize_endpoint": "https://auth.calendar-api.example.com/oauth/authorize"
    },
    "scopes": [
      {
        "scope_value": "calendar.read",
        "description": "Read calendar events"
      },
      {
        "scope_value": "calendar.write",
        "description": "Create and modify calendar events"
      },
      {
        "scope_value": "calendar.events.readonly",
        "description": "Read-only access to calendar events"
      }
    ]
  }')

HTTP_CODE=$(echo "$SERVICE_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$SERVICE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Calendar API service created${NC}"
    SERVICE_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "  Service ID: $SERVICE_ID"
elif [ "$HTTP_CODE" = "409" ]; then
    echo -e "${YELLOW}✓ Calendar API service already exists${NC}"
else
    echo -e "${RED}✗ Failed to create Calendar API service${NC}"
    echo "  HTTP Status: $HTTP_CODE"
    exit 1
fi
echo ""

# Create Email Service
echo "Creating Email API service..."
SERVICE_RESPONSE=$(curl -s -X POST "${ADMIN_API}/services" \
  -H "Content-Type: application/json" \
  -w "\n%{http_code}" \
  -d '{
    "display_name": "Email API",
    "client_id": "email-api-client",
    "client_secret": "email-secret-abcde",
    "issuer_uri": "https://auth.email-api.example.com",
    "discovery": {
      "enable_discovery": false
    },
    "endpoints": {
      "token_endpoint": "https://auth.email-api.example.com/oauth/token",
      "authorize_endpoint": "https://auth.email-api.example.com/oauth/authorize"
    },
    "scopes": [
      {
        "scope_value": "email.read",
        "description": "Read email messages"
      },
      {
        "scope_value": "email.send",
        "description": "Send email messages"
      },
      {
        "scope_value": "email.compose",
        "description": "Compose draft emails"
      }
    ]
  }')

HTTP_CODE=$(echo "$SERVICE_RESPONSE" | tail -n1)
RESPONSE_BODY=$(echo "$SERVICE_RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
    echo -e "${GREEN}✓ Email API service created${NC}"
    SERVICE_ID=$(echo "$RESPONSE_BODY" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)
    echo "  Service ID: $SERVICE_ID"
elif [ "$HTTP_CODE" = "409" ]; then
    echo -e "${YELLOW}✓ Email API service already exists${NC}"
else
    echo -e "${RED}✗ Failed to create Email API service${NC}"
    echo "  HTTP Status: $HTTP_CODE"
    exit 1
fi
echo ""

echo "========================================="
echo "Development Environment Setup Complete"
echo "========================================="
echo ""
echo "Summary:"
echo "  Agents created: 3"
echo "    - Weather Assistant"
echo "    - Task Manager Pro"
echo "    - OAuth2 Test Client (upstream-oauth2-client)"
echo "      * Service requirements: Mock OAuth2 Service (mandatory)"
echo ""
echo "  Services created: 4"
echo "    - Mock OAuth2 Service (Dev)"
echo "    - Weather API"
echo "    - Calendar API"
echo "    - Email API"
echo ""
echo "Next steps:"
echo "  1. Start upstream OAuth2 mock server:"
echo "     just mock-upstream-oauth2-start"
echo "  2. Start identity broker:"
echo "     just dev"
echo "  3. Test OAuth2 flow at http://localhost:3000"
echo ""
echo "You can now test the consent management UI with this sample data."
echo ""
