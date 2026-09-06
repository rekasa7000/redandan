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

// POST /api/v1/push/subscribe
// Registers a Web Push subscription for the authenticated user.
func (h *Handler) PushSubscribe(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body struct {
		Endpoint string `json:"endpoint" binding:"required"`
		Keys     struct {
			P256dh string `json:"p256dh" binding:"required"`
			Auth   string `json:"auth" binding:"required"`
		} `json:"keys" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub := models.PushSubscription{
		ID:        bson.NewObjectID(),
		UserID:    oid,
		Endpoint:  body.Endpoint,
		P256dh:    body.Keys.P256dh,
		Auth:      body.Keys.Auth,
		CreatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Upsert by endpoint so re-registering the same browser doesn't create duplicates.
	opts := options.UpdateOne().SetUpsert(true)
	_, err = h.col("push_subscriptions").UpdateOne(ctx,
		bson.M{"user_id": oid, "endpoint": body.Endpoint},
		bson.M{"$set": sub},
		opts,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not save subscription"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "subscribed"})
}

// GET /api/v1/notifications
// Returns in-app notification records for the authenticated user.
func (h *Handler) ListNotifications(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(50)

	cursor, err := h.col("notifications").Find(ctx, bson.M{"user_id": oid}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch notifications"})
		return
	}
	defer cursor.Close(ctx)

	var notifications []models.Notification
	if err := cursor.All(ctx, &notifications); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

// PATCH /api/v1/notifications/:id
// Marks a notification as read.
func (h *Handler) MarkNotificationRead(c *gin.Context) {
	userID, notifOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.col("notifications").UpdateOne(ctx,
		bson.M{"_id": notifOID, "user_id": userID},
		bson.M{"$set": bson.M{"read": true}},
	)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "marked as read"})
}
