#!/bin/bash

# run-middleware-test.sh - Script testing robust untuk Sprint 3 Middleware
# Usage: ./run-middleware-test.sh

# HAPUS set -e supaya script tetap jalan meski satu test gagal
# set -e  # <-- JANGAN PAKAI INI

BASE_URL="http://localhost:8080"
JWT_SECRET="test-secret-key-for-development-only"

# Warna untuk output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo "=========================================="
echo "🧪 Sprint 3 Middleware Testing Suite"
echo "=========================================="
echo ""

# Helper function untuk generate JWT test token (lebih robust)
generate_jwt() {
    local user_id=$1
    local role=$2
    
    echo -e "${BLUE}🔐 Generating JWT for user=$user_id, role=$role${NC}" >&2
    
    # JWT Header
    header=$(echo -n '{"alg":"HS256","typ":"JWT"}' | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')
    
    # JWT Payload
    exp=$(($(date +%s) + 3600))
    payload=$(echo -n "{\"sub\":\"$user_id\",\"role\":\"$role\",\"exp\":$exp}" | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')
    
    # Signature (HMAC-SHA256) - pakai printf untuk hindari newline issue
    signature=$(printf "%s" "$header.$payload" | openssl dgst -sha256 -hmac "$JWT_SECRET" -binary | base64 | tr -d '=' | tr '/+' '_-' | tr -d '\n')
    
    local token="$header.$payload.$signature"
    echo -e "${GREEN}✓ JWT generated: ${token:0:50}...${NC}" >&2
    echo "$token"
}

echo "📝 Generating Test Tokens..."
echo ""

# Generate test tokens dengan error handling
CUSTOMER_TOKEN=$(generate_jwt "customer-123" "customer")
if [ -z "$CUSTOMER_TOKEN" ]; then
    echo -e "${RED}❌ Gagal generate CUSTOMER_TOKEN${NC}"
    exit 1
fi

DRIVER_TOKEN=$(generate_jwt "driver-456" "driver")
if [ -z "$DRIVER_TOKEN" ]; then
    echo -e "${RED}❌ Gagal generate DRIVER_TOKEN${NC}"
    exit 1
fi

INVALID_TOKEN="invalid.token.here"

echo ""
echo "✅ Test Tokens Ready:"
echo "  Customer: ${CUSTOMER_TOKEN:0:50}..."
echo "  Driver: ${DRIVER_TOKEN:0:50}..."
echo ""

# Test counter
PASS=0
FAIL=0
TOTAL=0

# Helper function untuk test dengan TIMEOUT (FIXED VERSION)
test_endpoint() {
    local test_name=$1
    local expected_status=$2
    local curl_cmd=$3
    
    ((TOTAL++))
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo -e "${BLUE}🔍 Test #$TOTAL: $test_name${NC}"
    echo "   Command: $curl_cmd"
    echo ""
    
    local response
    
    # Jalankan curl (tanpa -w), simpan output mentah ke variabel
    response=$(eval "$curl_cmd" -s --max-time 5 2>&1)
    
    # Check if curl failed (e.g. timeout, connection refused)
    if [ $? -ne 0 ]; then
        echo -e "   ${RED}❌ CURL FAILED (timeout or connection error)${NC}"
        echo "   Response: $response"
        ((FAIL++))
        echo ""
        return 1
    fi
    
    # ✅ FIX: Extract status code dari BARIS PERTAMA (contoh: "HTTP/1.1 200 OK")
    # Ini 100% robust dan tidak peduli apakah JSON body punya trailing newline atau tidak.
    local status_code=$(echo "$response" | head -n 1 | awk '{print $2}')
    
    echo "   Status Code: $status_code"
    echo "   Response Body:"
    # Tampilkan response (skip baris pertama yang merupakan HTTP status line)
    echo "$response" | tail -n +2 | head -15 | sed 's/^/     /'
    echo ""
    
    if [ "$status_code" == "$expected_status" ]; then
        echo -e "   ${GREEN}✓ PASS${NC} - Expected: $expected_status, Got: $status_code"
        ((PASS++))
    else
        echo -e "   ${RED}✗ FAIL${NC} - Expected: $expected_status, Got: $status_code"
        ((FAIL++))
    fi
    echo ""
}

echo "=========================================="
echo "📋 Test Suite 1: Client Info Middleware"
echo "=========================================="
echo ""

test_endpoint \
    "Client Info - All Headers (Debug Build)" \
    "200" \
    "curl -i $BASE_URL/ping \
    -H 'X-App-Version: 1.2.0' \
    -H 'X-Platform: android' \
    -H 'X-OS-Version: 14' \
    -H 'X-Build: debug'"

test_endpoint \
    "Client Info - No Headers (Fail Open)" \
    "200" \
    "curl -i $BASE_URL/ping"

echo "=========================================="
echo "📋 Test Suite 2: JWT Auth Middleware (REST)"
echo "=========================================="
echo ""

test_endpoint \
    "REST Auth - Valid Customer Token" \
    "200" \
    "curl -i $BASE_URL/api/v1/orders/history \
    -H 'Authorization: Bearer $CUSTOMER_TOKEN'"

test_endpoint \
    "REST Auth - Missing Authorization Header" \
    "401" \
    "curl -i $BASE_URL/api/v1/orders/history"

test_endpoint \
    "REST Auth - Invalid JWT Format" \
    "401" \
    "curl -i $BASE_URL/api/v1/orders/history \
    -H 'Authorization: Bearer invalid.token.here'"

echo "=========================================="
echo "📋 Test Suite 3: WebSocket Auth Middleware"
echo "=========================================="
echo ""

test_endpoint \
    "WS Auth - Valid Token (Customer)" \
    "200" \
    "curl -i '$BASE_URL/v1/ws/chat/order-123?token=$CUSTOMER_TOKEN'"

test_endpoint \
    "WS Auth - Missing Token Parameter" \
    "401" \
    "curl -i '$BASE_URL/v1/ws/chat/order-123'"

test_endpoint \
    "WS Auth - Invalid Token" \
    "401" \
    "curl -i '$BASE_URL/v1/ws/chat/order-123?token=$INVALID_TOKEN'"

test_endpoint \
    "WS Auth - Role Enforcement (Driver endpoint with Customer token)" \
    "403" \
    "curl -i '$BASE_URL/v1/ws/driver/location?token=$CUSTOMER_TOKEN'"

echo "=========================================="
echo "📊 Test Summary"
echo "=========================================="
echo ""
echo -e "${GREEN}✓ Passed: $PASS${NC}"
echo -e "${RED}✗ Failed: $FAIL${NC}"
echo ""
echo "Total: $TOTAL tests"
echo ""

if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    exit 0
else
    echo -e "${RED}❌ Some tests failed. Please review the output above.${NC}"
    exit 1
fi