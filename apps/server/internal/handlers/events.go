package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"redandan/server/internal/middleware"
	"redandan/server/internal/models"
)

// GET /api/v1/events
func (h *Handler) ListEvents(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "start_time", Value: 1}})
	cursor, err := h.col("events").Find(ctx, bson.M{"user_id": oid}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch events"})
		return
	}
	defer cursor.Close(ctx)

	var events []models.Event
	if err := cursor.All(ctx, &events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode events"})
		return
	}

	c.JSON(http.StatusOK, events)
}

// POST /api/v1/events
func (h *Handler) CreateEvent(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body struct {
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		StartTime   time.Time `json:"start_time" binding:"required"`
		EndTime     time.Time `json:"end_time"`
		AllDay      bool      `json:"all_day"`
		ContextID   string    `json:"context_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := models.Event{
		ID:          bson.NewObjectID(),
		UserID:      oid,
		Title:       body.Title,
		Description: body.Description,
		StartTime:   body.StartTime,
		EndTime:     body.EndTime,
		AllDay:      body.AllDay,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if body.ContextID != "" {
		cid, err := bson.ObjectIDFromHex(body.ContextID)
		if err == nil {
			event.ContextID = cid
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.col("events").InsertOne(ctx, event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create event"})
		return
	}

	c.JSON(http.StatusCreated, event)
}

// GET /api/v1/events/:id
func (h *Handler) GetEvent(c *gin.Context) {
	userID, evOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var event models.Event
	if err := h.col("events").FindOne(ctx, bson.M{"_id": evOID, "user_id": userID}).Decode(&event); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

// PATCH /api/v1/events/:id
func (h *Handler) UpdateEvent(c *gin.Context) {
	userID, evOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	var body struct {
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		StartTime   *time.Time `json:"start_time"`
		EndTime     *time.Time `json:"end_time"`
		AllDay      *bool      `json:"all_day"`
		ContextID   *string    `json:"context_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	set := bson.M{"updated_at": time.Now()}
	if body.Title != nil {
		set["title"] = *body.Title
	}
	if body.Description != nil {
		set["description"] = *body.Description
	}
	if body.StartTime != nil {
		set["start_time"] = *body.StartTime
	}
	if body.EndTime != nil {
		set["end_time"] = *body.EndTime
	}
	if body.AllDay != nil {
		set["all_day"] = *body.AllDay
	}
	if body.ContextID != nil {
		cid, err := bson.ObjectIDFromHex(*body.ContextID)
		if err == nil {
			set["context_id"] = cid
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.col("events").UpdateOne(ctx,
		bson.M{"_id": evOID, "user_id": userID},
		bson.M{"$set": set},
	)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	var updated models.Event
	h.col("events").FindOne(ctx, bson.M{"_id": evOID}).Decode(&updated)
	c.JSON(http.StatusOK, updated)
}

// DELETE /api/v1/events/:id
func (h *Handler) DeleteEvent(c *gin.Context) {
	userID, evOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.col("events").DeleteOne(ctx, bson.M{"_id": evOID, "user_id": userID})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
