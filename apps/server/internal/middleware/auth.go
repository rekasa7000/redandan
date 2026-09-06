package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	ContextKeyUserID    = "userID"
	ContextKeyTokenType = "tokenType"

	TokenTypeAccess  = "access"
	TokenTypePending = "pending" // issued after step-1 (password), before TOTP
)

// Claims is the JWT payload for both pending and access tokens.
type Claims struct {
	Type string `json:"type"` // "access" | "pending"
	jwt.RegisteredClaims
}

// RequireAuth validates the Bearer token and rejects requests without a valid
// access token. Pending tokens (step-1 only) are explicitly rejected here.
func RequireAuth(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := extractClaims(c, jwtSecret)
		if err != nil || claims.Type != TokenTypeAccess {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(ContextKeyUserID, claims.Subject)
		c.Set(ContextKeyTokenType, claims.Type)
		c.Next()
	}
}

// RequirePending validates the Bearer token and requires it to be a pending token.
// Used exclusively on POST /api/v1/auth/totp/validate (step 2 of login).
func RequirePending(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := extractClaims(c, jwtSecret)
		if err != nil || claims.Type != TokenTypePending {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "pending token required"})
			return
		}

		c.Set(ContextKeyUserID, claims.Subject)
		c.Next()
	}
}

// RequireCron validates the CRON_SECRET bearer token on cron-only endpoints.
func RequireCron(cronSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "Bearer "+cronSecret {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}
}

// extractClaims parses and validates the Bearer token from the Authorization header.
func extractClaims(c *gin.Context, jwtSecret string) (*Claims, error) {
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, jwt.ErrTokenMalformed
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	claims := &Claims{}

	_, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}))

	return claims, err
}
