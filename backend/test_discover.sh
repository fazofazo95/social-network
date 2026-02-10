#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Testing Discover Users Functionality${NC}"
echo "========================================"

# Test user credentials
EMAIL="alice@example.com"
PASSWORD="Password123!"

echo -e "\n${BLUE}Step 1: Logging in as alice...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -c cookies.txt \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

echo "$LOGIN_RESPONSE" | jq .

if echo "$LOGIN_RESPONSE" | jq -e '.status == "success"' > /dev/null; then
  echo -e "${GREEN}✓ Login successful${NC}"
else
  echo -e "${RED}✗ Login failed${NC}"
  echo "Make sure the server is running and test users exist."
  echo "You can seed users by running: go run tools/seed/seed.go"
  rm -f cookies.txt
  exit 1
fi

echo -e "\n${BLUE}Step 2: Fetching feed (includes discover users)...${NC}"
FEED_RESPONSE=$(curl -s -X GET http://localhost:8080/api/feed \
  -b cookies.txt)

echo "$FEED_RESPONSE" | jq .

# Check if discovered_users is present
if echo "$FEED_RESPONSE" | jq -e '.data.discovered_users' > /dev/null; then
  echo -e "\n${GREEN}✓ Discovered users data found${NC}"
  
  USER_COUNT=$(echo "$FEED_RESPONSE" | jq '.data.discovered_users | length')
  echo -e "${GREEN}Found $USER_COUNT discovered users${NC}"
  
  if [ "$USER_COUNT" -gt 0 ]; then
    echo -e "\n${BLUE}Discovered Users:${NC}"
    echo "$FEED_RESPONSE" | jq -r '.data.discovered_users[] | "  - \(.first_name) \(.last_name) (ID: \(.id)) - Status: \(.status)"'
  fi
else
  echo -e "${RED}✗ No discovered users data in response${NC}"
fi

# Check posts
if echo "$FEED_RESPONSE" | jq -e '.data.posts' > /dev/null; then
  POST_COUNT=$(echo "$FEED_RESPONSE" | jq '.data.posts | length')
  echo -e "\n${GREEN}Found $POST_COUNT posts from followed users${NC}"
else
  echo -e "\n${BLUE}Note: No posts found (expected if no users are being followed yet)${NC}"
fi

# Cleanup
rm -f cookies.txt
echo -e "\n${GREEN}Test complete!${NC}"
