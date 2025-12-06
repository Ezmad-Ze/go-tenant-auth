#!/bin/bash

# Quick test for backup code and rate limiting functionality

API_URL="http://localhost:6969"
TIMESTAMP=$(date +%s)
TENANT_SLUG="test_${TIMESTAMP}"
EMAIL="user_${TIMESTAMP}@example.com"
PASSWORD="SecurePass123!"

echo "=== Testing Backup Code and Rate Limiting ==="

# 1. Create tenant
echo -e "\n1. Creating tenant..."
TENANT_RESPONSE=$(curl -s -X POST "$API_URL/tenants" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test Tenant $TIMESTAMP\",
    \"slug\": \"$TENANT_SLUG\",
    \"admin_email\": \"admin_${TIMESTAMP}@example.com\",
    \"admin_password\": \"AdminPass123!\"
  }")

TENANT_ID=$(echo "$TENANT_RESPONSE" | jq -r '.tenant.id')
echo "Tenant ID: $TENANT_ID"

# 2. Register user
echo -e "\n2. Registering user..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

ACCESS_TOKEN=$(echo "$REGISTER_RESPONSE" | jq -r '.access_token')
echo "Registered and got token"

# 3. Setup 2FA
echo -e "\n3. Setting up 2FA..."
SETUP_RESPONSE=$(curl -s -X POST "$API_URL/auth/2fa/setup" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID")

SECRET=$(echo "$SETUP_RESPONSE" | jq -r '.secret')
BACKUP_CODE=$(echo "$SETUP_RESPONSE" | jq -r '.backup_codes[0]')
echo "Secret: $SECRET"
echo "First backup code: $BACKUP_CODE"

# Build totp-gen if not exists
if [ ! -f "bin/totp-gen" ]; then
  go build -o bin/totp-gen cmd/totp-gen/main.go
fi

# 4. Enable 2FA
echo -e "\n4. Enabling 2FA..."
CODE=$(./bin/totp-gen "$SECRET")
ENABLE_RESPONSE=$(curl -s -X POST "$API_URL/auth/2fa/enable" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"code\": \"$CODE\"
  }")
echo "Response: $ENABLE_RESPONSE"

# 5. Test backup code login
echo -e "\n5. Testing login with backup code..."
BACKUP_LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"backup_code\": \"$BACKUP_CODE\"
  }")

if echo "$BACKUP_LOGIN_RESPONSE" | jq -e '.access_token' > /dev/null; then
  echo "✅ Backup code login successful!"
else
  echo "❌ Backup code login failed: $BACKUP_LOGIN_RESPONSE"
fi

# 6. Test rate limiting (try invalid codes multiple times)
echo -e "\n6. Testing rate limiting (5 failed attempts)..."
for i in {1..6}; do
  echo "Attempt $i..."
  FAIL_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $TENANT_ID" \
    -d "{
      \"email\": \"$EMAIL\",
      \"password\": \"$PASSWORD\",
      \"totp\": \"000000\"
    }")
  
  if echo "$FAIL_RESPONSE" | grep -q "locked"; then
    echo "✅ Account locked after failed attempts!"
    break
  fi
done

echo -e "\n=== Test Complete ==="
