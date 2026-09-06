package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"redandan/server/internal/middleware"
	"redandan/server/internal/models"
)

// POST /api/v1/auth/login
// Step 1 of two-step login: verify password, return pending JWT.
func (h *Handler) Login(c *gin.Context) {
	var body struct {
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	err := h.col("users").FindOne(ctx, bson.M{}).Decode(&user)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// Issue a short-lived pending token; TOTP must be verified next.
	pendingToken, err := h.signToken(user.ID.Hex(), middleware.TokenTypePending, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"pending_token": pendingToken})
}

// POST /api/v1/auth/totp/validate
// Step 2 of two-step login: verify TOTP code, return access JWT.
// Requires RequirePending middleware.
func (h *Handler) TOTPValidate(c *gin.Context) {
	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := h.col("users").FindOne(ctx, bson.M{"_id": oid}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if !user.TOTPEnabled || user.TOTPSecret == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "TOTP not configured"})
		return
	}

	if !totp.Validate(body.Code, user.TOTPSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid TOTP code"})
		return
	}

	accessToken, err := h.signToken(userID, middleware.TokenTypeAccess, 30*24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

// GET /api/v1/auth/totp/setup
// Generate a new TOTP secret and return the otpauth:// URI + QR code data URL.
// Requires RequireAuth middleware (must be logged in to (re)set TOTP).
func (h *Handler) TOTPSetup(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := h.col("users").FindOne(ctx, bson.M{"_id": oid}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Redandan",
		AccountName: user.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate TOTP secret"})
		return
	}

	// Store the pending (unconfirmed) secret; overwrite any previous pending secret.
	_, err = h.col("users").UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"totp_pending_secret": key.Secret()}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save pending secret"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"uri":    key.URL(),
		"secret": key.Secret(), // displayed so user can enter manually
	})
}

// POST /api/v1/auth/totp/confirm
// Confirm TOTP setup by verifying the first code against the pending secret.
// Requires RequireAuth middleware.
func (h *Handler) TOTPConfirm(c *gin.Context) {
	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := h.col("users").FindOne(ctx, bson.M{"_id": oid}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if user.TOTPPendingSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no pending TOTP setup; call /totp/setup first"})
		return
	}

	if !totp.Validate(body.Code, user.TOTPPendingSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid TOTP code"})
		return
	}

	// Generate backup codes.
	backupCodes, hashes, err := generateBackupCodes(8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate backup codes"})
		return
	}

	_, err = h.col("users").UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{
			"totp_secret":         user.TOTPPendingSecret,
			"totp_enabled":        true,
			"totp_pending_secret": "",
			"backup_codes":        hashes,
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not activate TOTP"})
		return
	}

	// Return plaintext codes once; user must save them now.
	c.JSON(http.StatusOK, gin.H{
		"message":      "TOTP enabled",
		"backup_codes": backupCodes,
	})
}

// POST /api/v1/auth/backup-code
// Validate a backup code in place of a TOTP code. Burns the code on success.
// Requires RequirePending middleware (same step-2 position as TOTPValidate).
func (h *Handler) BackupCode(c *gin.Context) {
	var body struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code required"})
		return
	}

	userID := c.GetString(middleware.ContextKeyUserID)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var user models.User
	if err := h.col("users").FindOne(ctx, bson.M{"_id": oid}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Find and burn the matching backup code.
	matchIdx := -1
	for i, hash := range user.BackupCodes {
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.Code)); err == nil {
			matchIdx = i
			break
		}
	}
	if matchIdx == -1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid backup code"})
		return
	}

	remaining := append(user.BackupCodes[:matchIdx], user.BackupCodes[matchIdx+1:]...)
	_, err = h.col("users").UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"backup_codes": remaining}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not burn backup code"})
		return
	}

	accessToken, err := h.signToken(userID, middleware.TokenTypeAccess, 30*24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"access_token": accessToken})
}

// POST /api/v1/auth/logout
// Stateless: the client discards the token. Server always returns 200.
func (h *Handler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

// signToken creates and signs a JWT with the given subject, type, and TTL.
func (h *Handler) signToken(subject, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := middleware.Claims{
		Type: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.cfg.JWTSecret))
}

// generateBackupCodes returns n random plaintext codes and their bcrypt hashes.
func generateBackupCodes(n int) ([]string, []string, error) {
	codes := make([]string, n)
	hashes := make([]string, n)
	for i := range codes {
		b := make([]byte, 6)
		if _, err := rand.Read(b); err != nil {
			return nil, nil, err
		}
		codes[i] = hex.EncodeToString(b) // 12 hex chars
		hash, err := bcrypt.GenerateFromPassword([]byte(codes[i]), bcrypt.DefaultCost)
		if err != nil {
			return nil, nil, err
		}
		hashes[i] = string(hash)
	}
	return codes, hashes, nil
}
