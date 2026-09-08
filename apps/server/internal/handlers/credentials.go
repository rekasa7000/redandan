package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"

	"reliva/server/internal/middleware"
	"reliva/server/internal/models"
)

// GET /api/v1/credentials
// Returns encrypted credential records. Decryption happens client-side only.
func (h *Handler) ListCredentials(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := h.col("credentials").Find(ctx, bson.M{"user_id": oid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch credentials"})
		return
	}
	defer cursor.Close(ctx)

	var credentials []models.Credential
	if err := cursor.All(ctx, &credentials); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode credentials"})
		return
	}

	c.JSON(http.StatusOK, credentials)
}

// POST /api/v1/credentials
// Stores an encrypted credential. The server never sees plaintext passwords.
// The client must encrypt before sending; the server stores ciphertext + IV + salt.
func (h *Handler) CreateCredential(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body struct {
		Site            string `json:"site" binding:"required"`
		Username        string `json:"username"`
		EncryptedData   string `json:"encrypted_data" binding:"required"`
		IV              string `json:"iv" binding:"required"`
		Salt            string `json:"salt" binding:"required"`
		Notes           string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cred := models.Credential{
		ID:            bson.NewObjectID(),
		UserID:        oid,
		Site:          body.Site,
		Username:      body.Username,
		EncryptedData: body.EncryptedData,
		IV:            body.IV,
		Salt:          body.Salt,
		Notes:         body.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.col("credentials").InsertOne(ctx, cred); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create credential"})
		return
	}

	c.JSON(http.StatusCreated, cred)
}

// GET /api/v1/credentials/:id
func (h *Handler) GetCredential(c *gin.Context) {
	userID, credOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var cred models.Credential
	if err := h.col("credentials").FindOne(ctx, bson.M{"_id": credOID, "user_id": userID}).Decode(&cred); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
		return
	}

	c.JSON(http.StatusOK, cred)
}

// PATCH /api/v1/credentials/:id
func (h *Handler) UpdateCredential(c *gin.Context) {
	userID, credOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	var body struct {
		Site          *string `json:"site"`
		Username      *string `json:"username"`
		EncryptedData *string `json:"encrypted_data"`
		IV            *string `json:"iv"`
		Salt          *string `json:"salt"`
		Notes         *string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	set := bson.M{"updated_at": time.Now()}
	if body.Site != nil {
		set["site"] = *body.Site
	}
	if body.Username != nil {
		set["username"] = *body.Username
	}
	if body.EncryptedData != nil {
		set["encrypted_data"] = *body.EncryptedData
	}
	if body.IV != nil {
		set["iv"] = *body.IV
	}
	if body.Salt != nil {
		set["salt"] = *body.Salt
	}
	if body.Notes != nil {
		set["notes"] = *body.Notes
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.col("credentials").UpdateOne(ctx,
		bson.M{"_id": credOID, "user_id": userID},
		bson.M{"$set": set},
	)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
		return
	}

	var updated models.Credential
	h.col("credentials").FindOne(ctx, bson.M{"_id": credOID}).Decode(&updated)
	c.JSON(http.StatusOK, updated)
}

// DELETE /api/v1/credentials/:id
func (h *Handler) DeleteCredential(c *gin.Context) {
	userID, credOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.col("credentials").DeleteOne(ctx, bson.M{"_id": credOID, "user_id": userID})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "credential not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
