package handlers

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"

	"reliva/server/internal/middleware"
	"reliva/server/internal/models"
)

// POST /api/v1/auth/login
// Step 1 of two-step login: verify email + password.
// Returns pending token (TOTP enabled) or full access token (first-time setup).
func (h *Handler) Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email"    binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and password required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	if err := h.col("users").FindOne(ctx, bson.M{"email": body.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	// TOTP not yet configured — issue a full access token so the user can
	// complete TOTP setup in /settings before accessing the rest of the app.
	if !user.TOTPEnabled {
		accessToken, err := h.signToken(user.ID.Hex(), middleware.TokenTypeAccess, 30*24*time.Hour)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"token":        accessToken,
			"require_totp": false,
		})
		return
	}

	// TOTP configured — issue a short-lived pending token; step 2 required.
	pendingToken, err := h.signToken(user.ID.Hex(), middleware.TokenTypePending, 5*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"pending_token": pendingToken,
		"require_totp":  true,
	})
}

// GET /api/v1/auth/me
// Returns basic profile info for the authenticated user.
func (h *Handler) Me(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"email":        user.Email,
		"totp_enabled": user.TOTPEnabled,
	})
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

	c.JSON(http.StatusOK, gin.H{"token": accessToken})
}

// GET /api/v1/auth/totp/setup
// Generate a new TOTP secret and return a QR code PNG data URL + manual secret.
// Requires RequireAuth middleware.
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
		Issuer:      "Reliva",
		AccountName: user.Email,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate TOTP secret"})
		return
	}

	_, err = h.col("users").UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"totp_pending_secret": key.Secret()}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save pending secret"})
		return
	}

	// Generate QR as PNG data URL — frontend needs no QR library.
	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not generate QR code"})
		return
	}
	var buf bytes.Buffer
	buf.Write(png)
	qrDataURL := "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())

	c.JSON(http.StatusOK, gin.H{
		"qr":     qrDataURL,
		"secret": key.Secret(),
		"uri":    key.URL(),
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

	c.JSON(http.StatusOK, gin.H{
		"message":      "TOTP enabled",
		"backup_codes": backupCodes,
	})
}

// POST /api/v1/auth/backup-code
// Validate a backup code in place of TOTP. Burns the code on success.
// Requires RequirePending middleware.
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

	c.JSON(http.StatusOK, gin.H{"token": accessToken})
}

// POST /api/v1/auth/change-password
// Change the authenticated user's password. Requires RequireAuth middleware.
func (h *Handler) ChangePassword(c *gin.Context) {
	var body struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password"     binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "current_password and new_password (min 8 chars) required"})
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(body.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "current password is incorrect"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}

	_, err = h.col("users").UpdateOne(ctx,
		bson.M{"_id": oid},
		bson.M{"$set": bson.M{"password_hash": string(newHash)}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not update password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password updated"})
}

// POST /api/v1/auth/forgot-password
// Reset password using a backup code — no email OTP required.
// The backup code proves identity; it is burned on success.
func (h *Handler) ForgotPassword(c *gin.Context) {
	var body struct {
		Email       string `json:"email"        binding:"required"`
		BackupCode  string `json:"backup_code"  binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email, backup_code, and new_password (min 8 chars) required"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	if err := h.col("users").FindOne(ctx, bson.M{"email": body.Email}).Decode(&user); err != nil {
		// Same error regardless of whether email exists — prevent enumeration.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	matchIdx := -1
	for i, hash := range user.BackupCodes {
		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(body.BackupCode)); err == nil {
			matchIdx = i
			break
		}
	}
	if matchIdx == -1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not hash password"})
		return
	}

	remaining := append(user.BackupCodes[:matchIdx], user.BackupCodes[matchIdx+1:]...)
	_, err = h.col("users").UpdateOne(ctx,
		bson.M{"_id": user.ID},
		bson.M{"$set": bson.M{
			"password_hash": string(newHash),
			"backup_codes":  remaining,
		}},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password reset — log in with your new password"})
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
