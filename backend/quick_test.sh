#!/bin/bash

# Quick test
curl -s -X GET http://localhost:8080/api/feed -b /tmp/test_cookies.txt > /tmp/feed_response.json 2>&1
cat /tmp/feed_response.json
