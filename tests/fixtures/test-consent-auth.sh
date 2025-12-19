#!/bin/bash
# Test consent API authentication requirements
set -e

AGENT_ID="55d6ba4e-428c-418a-a7ed-06953ac25263"
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "Testing Consent API Authentication"
echo "========================================="
echo ""

# Test 1: Direct to backend WITH header (should work)
echo "Test 1: Backend (port 8000) WITH X-Remote-User header"
echo "------------------------------------------------------"
echo "GET /api/me:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -H "X-Remote-User: dev@example.com" http://localhost:8000/api/me)
if [ "$STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 200)${NC}"
fi

echo "GET /api/consent/agents:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -H "X-Remote-User: dev@example.com" http://localhost:8000/api/consent/agents)
if [ "$STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 200)${NC}"
fi

echo "GET /api/consent/agent/:id:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -H "X-Remote-User: dev@example.com" http://localhost:8000/api/consent/agent/${AGENT_ID})
if [ "$STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 200)${NC}"
fi

echo "GET /api/consent/agent/:id/grants:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null -H "X-Remote-User: dev@example.com" http://localhost:8000/api/consent/agent/${AGENT_ID}/grants)
if [ "$STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 200)${NC}"
fi

echo ""

# Test 2: Direct to backend WITHOUT header (should fail with 401)
echo "Test 2: Backend (port 8000) WITHOUT X-Remote-User header"
echo "---------------------------------------------------------"
echo "GET /api/me:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8000/api/me)
if [ "$STATUS" = "401" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS, correctly rejected)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 401)${NC}"
fi

echo "GET /api/consent/agents:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8000/api/consent/agents)
if [ "$STATUS" = "401" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS, correctly rejected)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 401)${NC}"
fi

echo "GET /api/consent/agent/:id:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8000/api/consent/agent/${AGENT_ID})
if [ "$STATUS" = "401" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS, correctly rejected)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 401)${NC}"
fi

echo "GET /api/consent/agent/:id/grants:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:8000/api/consent/agent/${AGENT_ID}/grants)
if [ "$STATUS" = "401" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS, correctly rejected)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 401)${NC}"
fi

echo ""

# Test 3: Through Vite proxy WITHOUT header (proxy should inject it)
echo "Test 3: Vite Proxy (port 3000) WITHOUT X-Remote-User header"
echo "------------------------------------------------------------"
echo "GET /api/me:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:3000/api/me)
if [ "$STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS, proxy injected header)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 200 - proxy should inject header)${NC}"
fi

echo "GET /api/consent/agents:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:3000/api/consent/agents)
if [ "$STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS, proxy injected header)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 200 - proxy should inject header)${NC}"
fi

echo "GET /api/consent/agent/:id:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:3000/api/consent/agent/${AGENT_ID})
if [ "$STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS, proxy injected header)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 200 - proxy should inject header)${NC}"
fi

echo "GET /api/consent/agent/:id/grants:"
STATUS=$(curl -s -w "%{http_code}" -o /dev/null http://localhost:3000/api/consent/agent/${AGENT_ID}/grants)
if [ "$STATUS" = "200" ]; then
    echo -e "${GREEN}✓ Pass (HTTP $STATUS, proxy injected header)${NC}"
else
    echo -e "${RED}✗ Fail (HTTP $STATUS, expected 200 - proxy should inject header)${NC}"
fi

echo ""
echo "========================================="
echo "Test Summary"
echo "========================================="
echo "All consent API endpoints now require authentication."
echo "Vite proxy automatically injects X-Remote-User header for development."
echo ""
