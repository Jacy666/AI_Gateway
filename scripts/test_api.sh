#!/bin/bash

# AI Gateway API Test Script
# This script demonstrates the basic API functionality

set -e

BASE_URL="${BASE_URL:-http://localhost:8080}"
USERNAME="${USERNAME:-testuser}"
PASSWORD="${PASSWORD:-testpass123}"

echo "=========================================="
echo "AI Gateway API Test"
echo "=========================================="
echo "Base URL: $BASE_URL"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 1. Health Check
echo -e "${YELLOW}1. Health Check${NC}"
echo "GET $BASE_URL/health"
curl -s -X GET "$BASE_URL/health" | jq '.'
echo ""
echo ""

# 2. Register User (optional, might fail if user exists)
echo -e "${YELLOW}2. Register User${NC}"
echo "POST $BASE_URL/api/v1/auth/register"
curl -s -X POST "$BASE_URL/api/v1/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\",\"email\":\"test@example.com\"}" \
  | jq '.'
echo ""
echo ""

# 3. Login to get JWT token
echo -e "${YELLOW}3. Login${NC}"
echo "POST $BASE_URL/api/v1/auth/login"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

echo "$LOGIN_RESPONSE" | jq '.'

TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.token')

if [ "$TOKEN" = "null" ] || [ -z "$TOKEN" ]; then
  echo -e "${YELLOW}Warning: Failed to get token. The service might not be running.${NC}"
  exit 1
fi

echo ""
echo -e "${GREEN}✓ Token obtained successfully${NC}"
echo ""

# 4. Submit a task
echo -e "${YELLOW}4. Submit Task${NC}"
echo "POST $BASE_URL/api/v1/tasks"
TASK_RESPONSE=$(curl -s -X POST "$BASE_URL/api/v1/tasks" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "text_classification",
    "payload": {
      "text": "This is a sample text for AI processing",
      "model": "bert-base-uncased"
    }
  }')

echo "$TASK_RESPONSE" | jq '.'

TASK_ID=$(echo "$TASK_RESPONSE" | jq -r '.data.task_id')

if [ "$TASK_ID" = "null" ] || [ -z "$TASK_ID" ]; then
  echo -e "${YELLOW}Warning: Failed to submit task${NC}"
  exit 1
fi

echo ""
echo -e "${GREEN}✓ Task submitted: $TASK_ID${NC}"
echo ""

# 5. Query task status (multiple times to see progression)
echo -e "${YELLOW}5. Query Task Status${NC}"
for i in {1..5}; do
  echo "Attempt $i: GET $BASE_URL/api/v1/tasks/$TASK_ID"
  TASK_STATUS=$(curl -s -X GET "$BASE_URL/api/v1/tasks/$TASK_ID" \
    -H "Authorization: Bearer $TOKEN")
  
  echo "$TASK_STATUS" | jq '.'
  
  STATUS=$(echo "$TASK_STATUS" | jq -r '.data.status')
  
  if [ "$STATUS" = "completed" ] || [ "$STATUS" = "failed" ]; then
    echo ""
    echo -e "${GREEN}✓ Task $STATUS${NC}"
    break
  fi
  
  if [ $i -lt 5 ]; then
    echo "Status: $STATUS, waiting 2 seconds..."
    sleep 2
  fi
  echo ""
done

# 6. Check gateway status
echo -e "${YELLOW}6. Gateway Status${NC}"
echo "GET $BASE_URL/api/v1/status"
curl -s -X GET "$BASE_URL/api/v1/status" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
echo ""

echo ""
echo "=========================================="
echo -e "${GREEN}✓ All tests completed!${NC}"
echo "=========================================="
