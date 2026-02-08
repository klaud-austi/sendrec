#!/bin/bash
# Quick test script for SendRec deployment

set -e

echo "=== SendRec Deployment Test ==="
echo ""

# Check if server is running
if ! curl -s http://localhost:8080/health > /dev/null 2>&1; then
    echo "❌ Server not running. Starting..."
    docker-compose up -d
    sleep 3
fi

# Test health endpoint
echo "Testing health endpoint..."
HEALTH=$(curl -s http://localhost:8080/health)
if echo "$HEALTH" | grep -q "ok"; then
    echo "✅ Health check passed: $HEALTH"
else
    echo "❌ Health check failed"
    exit 1
fi

# Test landing page
echo "Testing landing page..."
if curl -s http://localhost:8080/ | grep -q "SendRec"; then
    echo "✅ Landing page loads"
else
    echo "❌ Landing page failed"
    exit 1
fi

# Test waitlist API
echo "Testing waitlist API..."
TEST_EMAIL="test-$(date +%s)@example.com"
RESPONSE=$(curl -s -X POST http://localhost:8080/waitlist \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\"}")

if echo "$RESPONSE" | grep -q '"success":true'; then
    echo "✅ Waitlist API works: $RESPONSE"
else
    echo "❌ Waitlist API failed: $RESPONSE"
    exit 1
fi

# Test duplicate prevention
echo "Testing duplicate prevention..."
DUP_RESPONSE=$(curl -s -X POST http://localhost:8080/waitlist \
    -H "Content-Type: application/json" \
    -d "{\"email\":\"$TEST_EMAIL\"}")

if echo "$DUP_RESPONSE" | grep -q "already registered"; then
    echo "✅ Duplicate prevention works"
else
    echo "❌ Duplicate prevention failed"
    exit 1
fi

echo ""
echo "=== All tests passed! ==="
echo "Landing page: http://localhost:8080"
echo "Admin panel: http://localhost:8080/admin"
