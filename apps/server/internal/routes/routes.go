package routes

import (
	"github.com/gin-gonic/gin"

	"redandan/server/internal/config"
	"redandan/server/internal/handlers"
	"redandan/server/internal/middleware"
)

// Register mounts all API routes onto the given engine.
func Register(r *gin.Engine, h *handlers.Handler, cfg *config.Config) {
	r.Use(middleware.CORS(cfg.AllowedOrigins))
	r.Use(middleware.StripTrailingSlash())

	v1 := r.Group("/api/v1")

	// Auth — public
	auth := v1.Group("/auth")
	{
		auth.POST("/login", h.Login)
		auth.POST("/logout", h.Logout)

		// Step 2: requires pending token
		pending := auth.Group("", middleware.RequirePending(cfg.JWTSecret))
		{
			pending.POST("/totp/validate", h.TOTPValidate)
			pending.POST("/backup-code", h.BackupCode)
		}

		// TOTP management: requires full access token
		totpGroup := auth.Group("/totp", middleware.RequireAuth(cfg.JWTSecret))
		{
			totpGroup.GET("/setup", h.TOTPSetup)
			totpGroup.POST("/confirm", h.TOTPConfirm)
		}
	}

	// All routes below require a valid access token.
	protected := v1.Group("", middleware.RequireAuth(cfg.JWTSecret))
	{
		// Contexts
		protected.GET("/contexts", h.ListContexts)
		protected.POST("/contexts", h.CreateContext)
		protected.GET("/contexts/:id", h.GetContext)
		protected.PATCH("/contexts/:id", h.UpdateContext)
		protected.DELETE("/contexts/:id", h.DeleteContext)

		// Tasks
		protected.GET("/tasks", h.ListTasks)
		protected.POST("/tasks", h.CreateTask)
		protected.GET("/tasks/:id", h.GetTask)
		protected.PATCH("/tasks/:id", h.UpdateTask)
		protected.DELETE("/tasks/:id", h.DeleteTask)

		// Events
		protected.GET("/events", h.ListEvents)
		protected.POST("/events", h.CreateEvent)
		protected.GET("/events/:id", h.GetEvent)
		protected.PATCH("/events/:id", h.UpdateEvent)
		protected.DELETE("/events/:id", h.DeleteEvent)

		// Credentials (zero-knowledge vault)
		protected.GET("/credentials", h.ListCredentials)
		protected.POST("/credentials", h.CreateCredential)
		protected.GET("/credentials/:id", h.GetCredential)
		protected.PATCH("/credentials/:id", h.UpdateCredential)
		protected.DELETE("/credentials/:id", h.DeleteCredential)

		// Notifications
		protected.GET("/notifications", h.ListNotifications)
		protected.PATCH("/notifications/:id", h.MarkNotificationRead)

		// Push
		protected.POST("/push/subscribe", h.PushSubscribe)
	}

	// Cron — protected by CRON_SECRET bearer token, not a user JWT.
	cron := v1.Group("/cron", middleware.RequireCron(cfg.CronSecret))
	{
		cron.POST("/notify", h.CronNotify)
	}
}
