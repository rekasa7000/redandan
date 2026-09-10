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

// GET /api/v1/contexts
func (h *Handler) ListContexts(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := h.col("contexts").Find(ctx, bson.M{"user_id": oid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch contexts"})
		return
	}
	defer cursor.Close(ctx)

	contexts := []models.Context{}
	if err := cursor.All(ctx, &contexts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode contexts"})
		return
	}

	c.JSON(http.StatusOK, contexts)
}

// POST /api/v1/contexts
func (h *Handler) CreateContext(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Color       string `json:"color"`
		Icon        string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jobContext := models.Context{
		ID:          bson.NewObjectID(),
		UserID:      oid,
		Name:        body.Name,
		Description: body.Description,
		Color:       body.Color,
		Icon:        body.Icon,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.col("contexts").InsertOne(ctx, jobContext); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create context"})
		return
	}

	c.JSON(http.StatusCreated, jobContext)
}

// GET /api/v1/contexts/:id
func (h *Handler) GetContext(c *gin.Context) {
	userID, ctxOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var jobContext models.Context
	err := h.col("contexts").FindOne(ctx, bson.M{"_id": ctxOID, "user_id": userID}).Decode(&jobContext)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "context not found"})
		return
	}

	c.JSON(http.StatusOK, jobContext)
}

// PATCH /api/v1/contexts/:id
func (h *Handler) UpdateContext(c *gin.Context) {
	userID, ctxOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	var body struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
		Icon        *string `json:"icon"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	set := bson.M{"updated_at": time.Now()}
	if body.Name != nil {
		set["name"] = *body.Name
	}
	if body.Description != nil {
		set["description"] = *body.Description
	}
	if body.Color != nil {
		set["color"] = *body.Color
	}
	if body.Icon != nil {
		set["icon"] = *body.Icon
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.col("contexts").UpdateOne(ctx,
		bson.M{"_id": ctxOID, "user_id": userID},
		bson.M{"$set": set},
	)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "context not found"})
		return
	}

	var updated models.Context
	h.col("contexts").FindOne(ctx, bson.M{"_id": ctxOID}).Decode(&updated)
	c.JSON(http.StatusOK, updated)
}

// DELETE /api/v1/contexts/:id
func (h *Handler) DeleteContext(c *gin.Context) {
	userID, ctxOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.col("contexts").DeleteOne(ctx, bson.M{"_id": ctxOID, "user_id": userID})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "context not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
