#!/bin/bash

echo "=== Testing Discover Users Functionality ==="
echo ""
echo "Prerequisites:"
echo "1. Backend server should be running on http://localhost:8080"
echo "2. Database should be seeded with test users"
echo ""
echo "--- Test 1: Login as alice ---"
echo "curl -X POST http://localhost:8080/api/login \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -c cookies.txt \\"
echo "  -d '{\"email\":\"alice@example.com\",\"password\":\"Password123!\"}'"
echo ""

curl -X POST http://localhost:8080/api/login \
  -H 'Content-Type: application/json' \
  -c cookies.txt \
  -d '{"email":"alice@example.com","password":"Password123!"}'

echo ""
echo ""
echo "--- Test 2: Get Feed (includes discover users) ---"
echo "curl -X GET http://localhost:8080/api/feed -b cookies.txt"
echo ""

curl -X GET http://localhost:8080/api/feed -b cookies.txt | jq .

echo ""
echo ""
echo "--- Cleanup ---"
rm -f cookies.txt
echo "Cookies removed"
