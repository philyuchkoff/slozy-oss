#!/bin/bash

# Security Fix Verification Script for SLOzy
# This script tests the implemented security fixes

BASE_URL="http://localhost:8080"
echo "Testing SLOzy Security Fixes..."
echo "================================="
echo "Make sure the server is running on $BASE_URL"
echo ""

# Test 1: Path Traversal in SLO Creation
echo "1. Testing Path Traversal in SLO Creation..."
curl -X POST "$BASE_URL/api/slo/create" \
  -H "Content-Type: application/json" \
  -d '{
    "serviceName": "../../../etc/passwd",
    "sloName": "test",
    "target": 99.9,
    "metricQuery": "test"
  }' -w "Status: %{http_code}\n" -o /tmp/test1.json

# Test 2: Large Payload DoS Protection
echo ""
echo "2. Testing Large Payload DoS Protection..."
LARGE_PAYLOAD=$(jq -c --argjson n 100000 '.metricQuery = "a" * $n' <<< '{}')
curl -X POST "$BASE_URL/api/slo/create" \
  -H "Content-Type: application/json" \
  -d "$LARGE_PAYLOAD" -w "Status: %{http_code}\n" -o /tmp/test2.json

# Test 3: Data Endpoint Security
echo ""
echo "3. Testing Data Endpoint Security..."
# Try to access non-yaml file
curl -X GET "$BASE_URL/data/../config/config.go" -w "Status: %{http_code}\n" -o /tmp/test3.txt
echo "Expected: 400 (Bad Request) or 403 (Forbidden)"

# Test 4: Security Headers
echo ""
echo "4. Testing Security Headers..."
curl -X GET "$BASE_URL/health" -I -o /tmp/test4_headers.txt
echo "Checking for security headers:"
grep -E "(X-Content-Type-Options|X-Frame-Options|Content-Security-Policy)" /tmp/test4_headers.txt || echo "WARNING: Missing security headers!"

# Test 5: Rate Limiting
echo ""
echo "5. Testing Rate Limiting..."
echo "Making rapid requests (should be rate limited after threshold)..."
for i in {1..5}; do
  code=$(curl -X GET "$BASE_URL/api/slos" -s -o /dev/null -w "%{http_code}")
  echo "Request $i: Status $code"
done

# Test 6: ZIP Filename Sanitization
echo ""
echo "6. Testing ZIP Filename Sanitization..."
curl -X POST "$BASE_URL/api/slo/create" \
  -H "Content-Type: application/json" \
  -d '{
    "serviceName": "test/service",
    "sloName": "../../../dangerous",
    "target": 99.9,
    "metricQuery": "test_metric"
  }' -o /tmp/test6.zip
file /tmp/test6.zip
unzip -l /tmp/test6.zip 2>/dev/null | grep -v "/"

# Test 7: CORS Headers
echo ""
echo "7. Testing CORS Headers..."
curl -X OPTIONS "$BASE_URL/api/slo/create" \
  -H "Origin: http://evil.com" \
  -H "Access-Control-Request-Method: POST" \
  -I -o /tmp/test7_cors.txt
echo "Access-Control-Allow-Origin:"
grep "Access-Control-Allow-Origin" /tmp/test7_cors.txt || echo "WARNING: Missing CORS headers!"

echo ""
echo "================================="
echo "Security Test Summary:"
echo "1. Path Traversal: $(if grep -q 'Service name is required' /tmp/test1.json; then echo "BLOCKED ✓"; else echo "VULNERABLE ✗"; fi)"
echo "2. Large Payload: $(if grep -q 'too large' /tmp/test2.json; then echo "BLOCKED ✓"; else echo "VULNERABLE ✗"; fi)"
echo "3. Data Endpoint: $(grep -q '400\|403' <<< "$(curl -s -o /dev/null -w "%{http_code}" "$BASE_URL/data/../config/config.go")" && echo "SECURED ✓" || echo "VULNERABLE ✗")"
echo "4. Security Headers: $(grep -q "X-Frame-Options" /tmp/test4_headers.txt && echo "PRESENT ✓" || echo "MISSING ✗")"
echo "5. Rate Limiting: $(tail -1 /tmp/rate_test.txt 2>/dev/null | grep -q "429" && echo "ACTIVE ✓" || echo "NOT TESTED")"
echo "6. ZIP Sanitization: $(zipinfo /tmp/test6.zip 2>/dev/null | grep -q "\.\./" && echo "VULNERABLE ✗" || echo "SECURED ✓")"
echo "7. CORS Protection: $(grep -q "http://evil.com" /tmp/test7_cors.txt && echo "VULNERABLE ✗" || echo "SECURED ✓")"

echo ""
echo "Cleaning up temp files..."
rm -f /tmp/test*.json /tmp/test*.txt /tmp/test*.zip

echo ""
echo "Done! Check the results above."