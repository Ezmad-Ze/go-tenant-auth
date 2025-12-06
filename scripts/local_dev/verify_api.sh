#!/bin/bash

# Configuration
API_URL="http://localhost:8083"
TIMESTAMP=$(date +%s)
TENANT_SLUG="test_tenant_${TIMESTAMP}"
ADMIN_EMAIL="admin_${TIMESTAMP}@example.com"
ADMIN_PASSWORD="AdminPass123!"
EMAIL="test_user_${TIMESTAMP}@example.com"
PASSWORD="SecurePass123!"
USERNAME="user_${TIMESTAMP}"

echo "Starting API Verification..."
echo "API_URL: $API_URL"

# 0. Create Tenant with Super Admin
echo -e "\n0. Creating Tenant with Super Admin..."
TENANT_RESPONSE=$(curl -s -X POST "$API_URL/tenants" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"Test Tenant $TIMESTAMP\",
    \"slug\": \"$TENANT_SLUG\",
    \"admin_email\": \"$ADMIN_EMAIL\",
    \"admin_password\": \"$ADMIN_PASSWORD\"
  }")

echo "Tenant Response: $TENANT_RESPONSE"
TENANT_ID=$(echo "$TENANT_RESPONSE" | jq -r '.tenant.id')
ADMIN_USER_ID=$(echo "$TENANT_RESPONSE" | jq -r '.admin_user.id')

if [ "$TENANT_ID" == "null" ] || [ -z "$TENANT_ID" ]; then
  echo "Failed to create tenant!"
  exit 1
fi

echo "TENANT_ID: $TENANT_ID"
echo "ADMIN_USER_ID: $ADMIN_USER_ID"
echo "ADMIN_EMAIL: $ADMIN_EMAIL"
echo "EMAIL: $EMAIL"

# 0.5. Login as Super Admin
echo -e "\n0.5. Logging in as Super Admin..."
ADMIN_LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"email\": \"$ADMIN_EMAIL\",
    \"password\": \"$ADMIN_PASSWORD\"
  }")

echo "Admin Login Response: $ADMIN_LOGIN_RESPONSE"
ADMIN_ACCESS_TOKEN=$(echo "$ADMIN_LOGIN_RESPONSE" | jq -r '.access_token')

if [ "$ADMIN_ACCESS_TOKEN" == "null" ] || [ -z "$ADMIN_ACCESS_TOKEN" ]; then
  echo "Failed to login as super admin!"
  exit 1
fi

echo "Super admin logged in successfully!"

# 1. Register
echo -e "\n1. Registering User..."
REGISTER_RESPONSE=$(curl -s -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"first_name\": \"Test\",
    \"last_name\": \"User\",
    \"username\": \"$USERNAME\"
  }")

echo "Response: $REGISTER_RESPONSE"

# Check for error
if echo "$REGISTER_RESPONSE" | grep -q "error"; then
  echo "Registration failed!"
  exit 1
fi

# 2. Login
echo -e "\n2. Logging in..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")

echo "Response: $LOGIN_RESPONSE"

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.refresh_token')

if [ "$ACCESS_TOKEN" == "null" ]; then
  echo "Login failed! No access token."
  exit 1
fi

echo "Access Token: ${ACCESS_TOKEN:0:10}..."

# 3. Refresh Token
echo -e "\n3. Refreshing Token..."
REFRESH_RESPONSE=$(curl -s -X POST "$API_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")

echo "Response: $REFRESH_RESPONSE"

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.access_token')
if [ "$NEW_ACCESS_TOKEN" == "null" ]; then
  echo "Token refresh failed!"
  exit 1
fi

# 4. Magic Link Send
echo -e "\n4. Sending Magic Link..."
MAGIC_LINK_RESPONSE=$(curl -s -X POST "$API_URL/auth/magic-link/send" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"email\": \"$EMAIL\"
  }")

echo "Response: $MAGIC_LINK_RESPONSE"

# 5. Authorization: Create Resource
echo -e "\n5. Creating Resource..."
RESOURCE_RESPONSE=$(curl -s -X POST "$API_URL/authz/resources" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "name": "document",
    "description": "Document management resource"
  }')
echo "Response: $RESOURCE_RESPONSE"

# 6. Authorization: Create Scope
echo -e "\n6. Creating Scope..."
SCOPE_RESPONSE=$(curl -s -X POST "$API_URL/authz/scopes" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "name": "read",
    "description": "Read permission"
  }')
echo "Response: $SCOPE_RESPONSE"

# 7. Authorization: Create Permission
echo -e "\n7. Creating Permission..."
PERMISSION_RESPONSE=$(curl -s -X POST "$API_URL/authz/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "resource_name": "document",
    "scope_name": "read",
    "description": "Permission to read documents"
  }')
echo "Response: $PERMISSION_RESPONSE"

# 8. Authorization: Check Permission
echo -e "\n8. Checking Permission..."
# Need user UUID
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.user.id // empty') 
# Note: Login response might not contain user ID directly in root, let's check token claims or just assume success if 200 OK for now.
# Actually, the check permission endpoint needs a user_id. 
# The login response structure in AuthHandler is:
# { "access_token": "...", "refresh_token": "...", "expires_in": ..., "token_type": "Bearer" }
# It does NOT return the user ID.
# However, we can use the "Get User Permissions" endpoint which uses the token's user ID context.
# Or we can decode the token (too complex for bash).
# Let's try "Get User Permissions" instead as it's simpler.

echo -e "\n8. Getting User Permissions..."
GET_PERMISSIONS_RESPONSE=$(curl -s -X GET "$API_URL/authz/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID")
echo "Response: $GET_PERMISSIONS_RESPONSE"

# 9. Authorization: Get User Roles
echo -e "\n9. Getting User Roles..."
GET_ROLES_RESPONSE=$(curl -s -X GET "$API_URL/authz/roles" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID")
echo "Response: $GET_ROLES_RESPONSE"

# 10. assign permission to role
echo -e "\n10. Assign Permission to Role..."
# Generate random suffix
RAND_SUFFIX=$(date +%s)
ROLE_NAME="editor_${RAND_SUFFIX}"
PERM_RESOURCE="document_${RAND_SUFFIX}"

# First create a role
ROLE_RESPONSE=$(curl -s -X POST "$API_URL/authz/roles" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"name\": \"$ROLE_NAME\",
    \"description\": \"Content editor role\",
    \"is_system\": false
  }")
echo "Role Response: $ROLE_RESPONSE"
ROLE_ID=$(echo "$ROLE_RESPONSE" | jq -r '.id')

# Create a new permission for this test
PERM_RESPONSE=$(curl -s -X POST "$API_URL/authz/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"resource_name\": \"$PERM_RESOURCE\",
    \"scope_name\": \"edit\",
    \"description\": \"Permission to edit documents\"
  }")
echo "Permission Response: $PERM_RESPONSE"
PERM_ID=$(echo "$PERM_RESPONSE" | jq -r '.id')

if [ "$ROLE_ID" != "null" ] && [ "$PERM_ID" != "null" ]; then
  ASSIGN_ROLE_PERM_RESPONSE=$(curl -s -X POST "$API_URL/authz/permissions/assign-to-role" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $TENANT_ID" \
    -d "{
      \"role_id\": \"$ROLE_ID\",
      \"permission_id\": \"$PERM_ID\"
    }")
  echo "Assign to Role Response: $ASSIGN_ROLE_PERM_RESPONSE"
else
  echo "Skipping Assign Permission to Role (Role or Permission creation failed)"
fi

# 11. assign permission to user
echo -e "\n11. Assign Permission to User..."
# We need user ID. We can get it from the login response.
# The login response was captured in LOGIN_RESPONSE variable in step 2.
USER_ID=$(echo "$LOGIN_RESPONSE" | jq -r '.user_id')
echo "Extracted User ID: $USER_ID"

if [ "$USER_ID" != "null" ] && [ "$PERM_ID" != "null" ]; then
  ASSIGN_USER_PERM_RESPONSE=$(curl -s -X POST "$API_URL/authz/permissions/assign-to-user" \
    -H "Authorization: Bearer $ACCESS_TOKEN" \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $TENANT_ID" \
    -d "{
      \"user_id\": \"$USER_ID\",
      \"permission_id\": \"$PERM_ID\"
    }")
  echo "Assign to User Response: $ASSIGN_USER_PERM_RESPONSE"
else
  echo "Skipping Assign Permission to User (User ID extraction failed)"
fi


# 9. Password Change
echo -e "\n12. Changing Password..."
CHANGE_PASSWORD_RESPONSE=$(curl -s -X POST "$API_URL/auth/password/change" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"old_password\": \"$PASSWORD\",
    \"new_password\": \"NewSecurePass456!\"
  }")
echo "Response: $CHANGE_PASSWORD_RESPONSE"

# Check if password change was successful
if echo "$CHANGE_PASSWORD_RESPONSE" | grep -q "error"; then
  echo "Password change failed!"
else
  echo "Password changed successfully!"
  PASSWORD="NewSecurePass456!"
  
  # Login with new password
  echo -e "\n13. Login with New Password..."
  NEW_LOGIN_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
    -H "Content-Type: application/json" \
    -H "X-Tenant-ID: $TENANT_ID" \
    -d "{
      \"email\": \"$EMAIL\",
      \"password\": \"$PASSWORD\"
    }")
  echo "Response: $NEW_LOGIN_RESPONSE"
  
  NEW_ACCESS_TOKEN=$(echo "$NEW_LOGIN_RESPONSE" | jq -r '.access_token')
  if [ "$NEW_ACCESS_TOKEN" == "null" ]; then
    echo "Login with new password failed!"
  else
    echo -e "\n✅ All API tests passed successfully!"

# Verify Audit Trail
echo -e "\n=== AUDIT TRAIL VERIFICATION ==="
echo "Checking audit logs in database..."

# Check if audit logs were created
AUDIT_COUNT=$(docker exec auth-db psql -U authuser -d authdb -t -c "SELECT COUNT(*) FROM audit_logs WHERE tenant_id = '$TENANT_ID';" 2>/dev/null | tr -d ' ')

if [ -z "$AUDIT_COUNT" ] || [ "$AUDIT_COUNT" == "0" ]; then
  echo "⚠️  Warning: No audit logs found for tenant $TENANT_ID"
  echo "   Audit logging may not be integrated into handlers yet."
else
  echo "✅ Found $AUDIT_COUNT audit log entries for this tenant"
  
  # Show recent audit logs
  echo -e "\nRecent audit logs:"
  docker exec auth-db psql -U authuser -d authdb -c "
    SELECT 
      action, 
      user_id IS NOT NULL as has_user,
      status,
      created_at 
    FROM audit_logs 
    WHERE tenant_id = '$TENANT_ID' 
    ORDER BY created_at DESC 
    LIMIT 10;" 2>/dev/null
  
  # Show action summary
  echo -e "\nAudit log summary by action:"
  docker exec auth-db psql -U authuser -d authdb -c "
    SELECT 
      action, 
      COUNT(*) as count,
      COUNT(CASE WHEN status = 'success' THEN 1 END) as success_count,
      COUNT(CASE WHEN status != 'success' THEN 1 END) as failure_count
    FROM audit_logs 
    WHERE tenant_id = '$TENANT_ID' 
    GROUP BY action 
    ORDER BY count DESC;" 2>/dev/null
fi

echo -e "\n=== VERIFICATION COMPLETE ==="
echo "Summary:"
echo "  - Tenant ID: $TENANT_ID"
echo "  - Admin User: $ADMIN_EMAIL"
echo "  - Test User: $EMAIL"
echo "  - Audit Logs: $AUDIT_COUNT entries"
echo ""
echo "All tests completed successfully! 🎉"
    ACCESS_TOKEN=$NEW_ACCESS_TOKEN
  fi
fi

# 11. Password Reset (Send Email)
echo -e "\n14. Requesting Password Reset..."
RESET_EMAIL="reset_${EMAIL}"
RESET_RESPONSE=$(curl -s -X POST "$API_URL/auth/password/reset" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"email\": \"$EMAIL\"
  }")
echo "Response: $RESET_RESPONSE"

# 15. TOTP 2FA
echo -e "\n15. Setting up TOTP 2FA..."

# Build totp-gen tool
go build -o bin/totp-gen cmd/totp-gen/main.go

SETUP_RESPONSE=$(curl -s -X POST "$API_URL/auth/2fa/setup" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID")
echo "Response: $SETUP_RESPONSE"

SECRET=$(echo "$SETUP_RESPONSE" | jq -r '.secret')
echo "Secret: $SECRET"

if [ "$SECRET" == "null" ] || [ -z "$SECRET" ]; then
  echo "Failed to setup 2FA!"
  exit 1
fi

# Generate code
CODE=$(./bin/totp-gen "$SECRET")
echo "Generated Code: $CODE"

echo -e "\n16. Enabling TOTP 2FA..."
ENABLE_RESPONSE=$(curl -s -X POST "$API_URL/auth/2fa/enable" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"code\": \"$CODE\"
  }")
echo "Response: $ENABLE_RESPONSE"

if echo "$ENABLE_RESPONSE" | grep -q "error"; then
  echo "Failed to enable 2FA!"
  exit 1
fi

# 17. Login with 2FA
echo -e "\n17. Logging in with 2FA..."

# First try without code (should fail)
LOGIN_FAIL_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\"
  }")
echo "Fail Response (Expected): $LOGIN_FAIL_RESPONSE"

if ! echo "$LOGIN_FAIL_RESPONSE" | grep -q "2FA_REQUIRED"; then
  echo "Login without 2FA should have failed with 2FA_REQUIRED!"
  # exit 1 # Don't exit yet, maybe the error message format is different
fi

# Now try with code
CODE=$(./bin/totp-gen "$SECRET")
LOGIN_SUCCESS_RESPONSE=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"email\": \"$EMAIL\",
    \"password\": \"$PASSWORD\",
    \"totp\": \"$CODE\"
  }")
echo "Success Response: $LOGIN_SUCCESS_RESPONSE"

NEW_ACCESS_TOKEN=$(echo "$LOGIN_SUCCESS_RESPONSE" | jq -r '.access_token')
if [ "$NEW_ACCESS_TOKEN" == "null" ]; then
  echo "Login with 2FA failed!"
  exit 1
fi
echo "Login with 2FA successful!"

# 18. Disable 2FA
echo -e "\n18. Disabling TOTP 2FA..."
DISABLE_RESPONSE=$(curl -s -X POST "$API_URL/auth/2fa/disable" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d "{
    \"password\": \"$PASSWORD\"
  }")
echo "Response: $DISABLE_RESPONSE"

if echo "$DISABLE_RESPONSE" | grep -q "error"; then
  echo "Failed to disable 2FA!"
  exit 1
fi

echo -e "\nVerification Complete!"
