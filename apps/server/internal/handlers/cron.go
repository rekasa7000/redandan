package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"reliva/server/internal/models"
)

// POST /api/v1/cron/notify
// Triggered by an external cron scheduler (Vercel, Railway, etc.).
// Finds tasks due today and creates in-app notification records.
// The endpoint is protected by RequireCron middleware (CRON_SECRET bearer token).
func (h *Handler) CronNotify(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	// Find all pending tasks due today across all users.
	filter := bson.M{
		"status":   bson.M{"$nin": []string{"done", "cancelled"}},
		"due_date": bson.M{"$gte": startOfDay, "$lt": endOfDay},
	}

	cursor, err := h.col("tasks").Find(ctx, filter, options.Find().SetProjection(bson.M{
		"_id":     1,
		"user_id": 1,
		"title":   1,
	}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not query tasks"})
		return
	}
	defer cursor.Close(ctx)

	var tasks []models.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not decode tasks"})
		return
	}

	if len(tasks) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "no tasks due today", "created": 0})
		return
	}

	notifications := make([]any, 0, len(tasks))
	for _, task := range tasks {
		notifications = append(notifications, models.Notification{
			ID:        bson.NewObjectID(),
			UserID:    task.UserID,
			Title:     "Task due today",
			Body:      task.Title,
			Read:      false,
			CreatedAt: now,
		})
	}

	result, err := h.col("notifications").InsertMany(ctx, notifications)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "notifications created",
		"created": len(result.InsertedIDs),
	})
}
