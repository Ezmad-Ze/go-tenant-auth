package security

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/o1egl/paseto"
)

// TokenClaims represents the claims in a token
type TokenClaims struct {
	UserID    uuid.UUID              `json:"user_id"`
	TenantID  uuid.UUID              `json:"tenant_id"`
	Email     string                 `json:"email"`
	Roles     []string               `json:"roles,omitempty"`
	TokenType string                 `json:"token_type"` // "access" or "refresh"
	IssuedAt  time.Time              `json:"iat"`
	ExpiresAt time.Time              `json:"exp"`
	Issuer    string                 `json:"iss"`
	Subject   string                 `json:"sub"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

// PasetoToken handles PASETO token operations
type PasetoToken struct {
	symmetricKey []byte
	issuer       string
}

// NewPasetoToken creates a new PASETO token handler
func NewPasetoToken(symmetricKey, issuer string) (*PasetoToken, error) {
	if len(symmetricKey) != 32 {
		return nil, fmt.Errorf("symmetric key must be exactly 32 bytes")
	}
	return &PasetoToken{
		symmetricKey: []byte(symmetricKey),
		issuer:       issuer,
	}, nil
}

// CreateToken creates a new PASETO token
func (p *PasetoToken) CreateToken(claims *TokenClaims, duration time.Duration) (string, error) {
	now := time.Now()
	claims.IssuedAt = now
	claims.ExpiresAt = now.Add(duration)
	claims.Issuer = p.issuer
	claims.Subject = claims.UserID.String()

	jsonToken := paseto.JSONToken{
		Audience:   p.issuer,
		Issuer:     p.issuer,
		Jti:        uuid.New().String(),
		Subject:    claims.Subject,
		IssuedAt:   claims.IssuedAt,
		Expiration: claims.ExpiresAt,
		NotBefore:  now,
	}

	// Add custom claims - using string representation for simplicity
	jsonToken.Set("user_id", claims.UserID.String())
	jsonToken.Set("tenant_id", claims.TenantID.String())
	jsonToken.Set("email", claims.Email)
	jsonToken.Set("token_type", claims.TokenType)
	// Roles stored as comma-separated string
	if len(claims.Roles) > 0 {
		jsonToken.Set("roles", strings.Join(claims.Roles, ","))
	}
	// Extra fields as individual key-value pairs
	if len(claims.Extra) > 0 {
		for k, v := range claims.Extra {
			jsonToken.Set(k, fmt.Sprint(v))
		}
	}

	v2 := paseto.NewV2()
	return v2.Encrypt(p.symmetricKey, jsonToken, nil)
}

// VerifyToken verifies and parses a PASETO token
func (p *PasetoToken) VerifyToken(token string) (*TokenClaims, error) {
	var jsonToken paseto.JSONToken
	var footer string

	v2 := paseto.NewV2()
	err := v2.Decrypt(token, p.symmetricKey, &jsonToken, &footer)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt token: %w", err)
	}

	// Verify expiration
	if time.Now().After(jsonToken.Expiration) {
		return nil, fmt.Errorf("token has expired")
	}

	// Extract claims
	claims := &TokenClaims{
		IssuedAt:  jsonToken.IssuedAt,
		ExpiresAt: jsonToken.Expiration,
		Issuer:    jsonToken.Issuer,
		Subject:   jsonToken.Subject,
	}

	// Parse custom claims - Get returns string value directly
	if userIDStr := jsonToken.Get("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			claims.UserID = userID
		}
	}

	if tenantIDStr := jsonToken.Get("tenant_id"); tenantIDStr != "" {
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			claims.TenantID = tenantID
		}
	}

	if email := jsonToken.Get("email"); email != "" {
		claims.Email = email
	}

	if tokenType := jsonToken.Get("token_type"); tokenType != "" {
		claims.TokenType = tokenType
	}

	// Parse roles from comma-separated string
	if rolesStr := jsonToken.Get("roles"); rolesStr != "" {
		claims.Roles = strings.Split(rolesStr, ",")
	}

	return claims, nil
}

// HashToken creates a SHA-256 hash of a token for storage
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
