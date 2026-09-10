package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"reliva/server/internal/middleware"
	"reliva/server/internal/models"
)

// GET /api/v1/tasks
func (h *Handler) ListTasks(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": oid}

	// Optional filters from query params.
	if status := c.Query("status"); status != "" {
		filter["status"] = status
	}
	if contextID := c.Query("context_id"); contextID != "" {
		cid, err := bson.ObjectIDFromHex(contextID)
		if err == nil {
			filter["context_id"] = cid
		}
	}

	opts := options.Find().SetSort(bson.D{{Key: "due_date", Value: 1}})
	cursor, err := h.col("tasks").Find(ctx, filter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch tasks"})
		return
	}
	defer cursor.Close(ctx)

	tasks := []models.Task{}
	if err := cursor.All(ctx, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// POST /api/v1/tasks
func (h *Handler) CreateTask(c *gin.Context) {
	userID := c.GetString(middleware.ContextKeyUserID)
	oid, err := bson.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var body struct {
		ContextID   string    `json:"context_id"`
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		Priority    string    `json:"priority"`
		DueDate     *time.Time `json:"due_date"`
		Tags        []string  `json:"tags"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := models.Task{
		ID:          bson.NewObjectID(),
		UserID:      oid,
		Title:       body.Title,
		Description: body.Description,
		Status:      "todo",
		Priority:    body.Priority,
		DueDate:     body.DueDate,
		Tags:        body.Tags,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if body.ContextID != "" {
		cid, err := bson.ObjectIDFromHex(body.ContextID)
		if err == nil {
			task.ContextID = cid
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := h.col("tasks").InsertOne(ctx, task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// GET /api/v1/tasks/:id
func (h *Handler) GetTask(c *gin.Context) {
	userID, taskOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var task models.Task
	err := h.col("tasks").FindOne(ctx, bson.M{"_id": taskOID, "user_id": userID}).Decode(&task)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// PATCH /api/v1/tasks/:id
func (h *Handler) UpdateTask(c *gin.Context) {
	userID, taskOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	var body struct {
		Title       *string    `json:"title"`
		Description *string    `json:"description"`
		Status      *string    `json:"status"`
		Priority    *string    `json:"priority"`
		DueDate     *time.Time `json:"due_date"`
		Tags        []string   `json:"tags"`
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
	if body.Status != nil {
		set["status"] = *body.Status
		if *body.Status == "done" {
			now := time.Now()
			set["completed_at"] = now
		}
	}
	if body.Priority != nil {
		set["priority"] = *body.Priority
	}
	if body.DueDate != nil {
		set["due_date"] = *body.DueDate
	}
	if body.Tags != nil {
		set["tags"] = body.Tags
	}
	if body.ContextID != nil {
		cid, err := bson.ObjectIDFromHex(*body.ContextID)
		if err == nil {
			set["context_id"] = cid
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.col("tasks").UpdateOne(ctx,
		bson.M{"_id": taskOID, "user_id": userID},
		bson.M{"$set": set},
	)
	if err != nil || result.MatchedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	var updated models.Task
	h.col("tasks").FindOne(ctx, bson.M{"_id": taskOID}).Decode(&updated)
	c.JSON(http.StatusOK, updated)
}

// DELETE /api/v1/tasks/:id
func (h *Handler) DeleteTask(c *gin.Context) {
	userID, taskOID, ok := h.parseUserAndResourceID(c)
	if !ok {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := h.col("tasks").DeleteOne(ctx, bson.M{"_id": taskOID, "user_id": userID})
	if err != nil || result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// parseUserAndResourceID is a helper that extracts the userID from the JWT context
// and the resource ObjectID from the :id route param. Writes error responses if invalid.
func (h *Handler) parseUserAndResourceID(c *gin.Context) (bson.ObjectID, bson.ObjectID, bool) {
	userStr := c.GetString(middleware.ContextKeyUserID)
	userOID, err := bson.ObjectIDFromHex(userStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return bson.ObjectID{}, bson.ObjectID{}, false
	}

	resourceOID, err := bson.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return bson.ObjectID{}, bson.ObjectID{}, false
	}

	return userOID, resourceOID, true
}
