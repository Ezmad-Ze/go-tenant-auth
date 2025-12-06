# API Examples

Collection of example requests for the Auth Service API.

## Environment Setup

Set these variables for convenience:
```bash
export API_URL="http://localhost:6969"
export TENANT_ID="00000000-0000-0000-0000-000000000001"
```

## Authentication Endpoints

### 1. Register a New User

```bash
curl -X POST "$API_URL/auth/register" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123!",
    "first_name": "Alice",
    "last_name": "Smith",
    "username": "alice_smith"
  }'
```

### 2. Login

```bash
curl -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123!"
  }'
```

### 3. Refresh Token

```bash
curl -X POST "$API_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "v2.local.yyy..."
  }'
```

### 4. Logout

```bash
export ACCESS_TOKEN="v2.local.xxx..."

curl -X POST "$API_URL/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID"
```

### 5. Magic Link

```bash
curl -X POST "$API_URL/auth/magic-link/send" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "email": "alice@example.com"
  }'
```

### 6. Verify Magic Link

```bash
curl -X POST "$API_URL/auth/magic-link/verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "token": "magic-token-here"
  }'
```

### 7. Change Password (Stub)

```bash
curl -X POST "$API_URL/auth/password/change" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "old_password": "SecurePass123!",
    "new_password": "NewSecurePass456!"
  }'
```

### 8. Reset Password (Stub)

```bash
curl -X POST "$API_URL/auth/password/reset" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "email": "alice@example.com"
  }'
```

### 9. Verify Email (Stub)

```bash
curl -X POST "$API_URL/auth/email/verify" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "token": "verification-token-here"
  }'
```

## TOTP 2FA Endpoints

### 1. Setup 2FA

Initiates 2FA setup by generating a secret, QR code URL, and backup codes.

```bash
curl -X POST "$API_URL/auth/2fa/setup" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID"
```

**Response:**
```json
{
  "secret": "JBSWY3DPEHPK3PXP",
  "qr_code_url": "otpauth://totp/AuthService:user@example.com?algorithm=SHA1&digits=6&issuer=AuthService&period=30&secret=JBSWY3DPEHPK3PXP",
  "backup_codes": [
    "ABC123DEF456",
    "GHI789JKL012",
    "MNO345PQR678",
    "STU901VWX234",
    "YZA567BCD890",
    "EFG123HIJ456",
    "KLM789NOP012",
    "QRS345TUV678",
    "WXY901ZAB234",
    "CDE567FGH890"
  ]
}
```

### 2. Enable 2FA

Enables 2FA after verifying a TOTP code from the authenticator app.

```bash
curl -X POST "$API_URL/auth/2fa/enable" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "code": "123456"
  }'
```

### 3. Verify 2FA Code

Verifies a TOTP code without enabling (useful for testing).

```bash
curl -X POST "$API_URL/auth/2fa/verify" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "code": "123456"
  }'
```

### 4. Disable 2FA

Disables 2FA for the authenticated user (requires password verification).

```bash
curl -X POST "$API_URL/auth/2fa/disable" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "password": "SecurePass123!"
  }'
```

### 5. Generate New Backup Codes

Generates a new set of backup codes (invalidates old ones).

```bash
curl -X POST "$API_URL/auth/2fa/backup-codes" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID"
```

### 6. Login with 2FA

When 2FA is enabled, provide the TOTP code in the login request.

```bash
curl -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123!",
    "totp": "123456"
  }'
```

### 7. Login with Backup Code

Use a backup code when TOTP is unavailable (e.g., lost device).

```bash
curl -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "email": "alice@example.com",
    "password": "SecurePass123!",
    "backup_code": "ABC123DEF456"
  }'
```

## Authorization Endpoints

### 1. Create a Resource

```bash
curl -X POST "$API_URL/authz/resources" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "name": "document",
    "description": "Document management resource"
  }'
```

### 2. Create a Scope

```bash
curl -X POST "$API_URL/authz/scopes" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "name": "share",
    "description": "Share permission"
  }'
```

### 3. Create a Permission

```bash
curl -X POST "$API_URL/authz/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "resource_name": "document",
    "scope_name": "share",
    "description": "Permission to share documents"
  }'
```

### 4. Create a Role

```bash
curl -X POST "$API_URL/authz/roles" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "name": "editor",
    "description": "Content editor role",
    "is_system": false
  }'
```

### 5. Assign Permission to Role

```bash
curl -X POST "$API_URL/authz/permissions/assign-to-role" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "role_id": "role-uuid-here",
    "permission_id": "permission-uuid-here"
  }'
```

### 6. Assign Permission to User

```bash
curl -X POST "$API_URL/authz/permissions/assign-to-user" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "user_id": "user-uuid-here",
    "permission_id": "permission-uuid-here"
  }'
```

### 7. Assign Role to User

```bash
curl -X POST "$API_URL/authz/roles/assign" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "user_id": "user-uuid-here",
    "role_id": "role-uuid-here"
  }'
```

### 8. Check Permission (RBAC)

```bash
curl -X POST "$API_URL/authz/check" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "user_id": "user-uuid-here",
    "resource": "document",
    "scope": "read"
  }'
```

### 9. Check Permission with ABAC Attributes

```bash
curl -X POST "$API_URL/authz/check" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -H "X-Tenant-ID: $TENANT_ID" \
  -d '{
    "user_id": "user-uuid-here",
    "resource": "document",
    "scope": "update",
    "attributes": {
      "document.owner_id": "user-uuid-here",
      "document.status": "draft"
    }
  }'
```

### 10. Get User Permissions

```bash
curl -X GET "$API_URL/authz/permissions" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID"
```

### 11. Get User Roles

```bash
curl -X GET "$API_URL/authz/users/user-uuid-here/roles" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "X-Tenant-ID: $TENANT_ID"
```

## Health Check

```bash
curl http://localhost:6969/health
```
