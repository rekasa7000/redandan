package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that sets Access-Control headers for allowed origins.
// The web frontend (same origin) uses cookies; extension and mobile use bearer tokens
// from their respective origins — both need CORS.
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin != "" && slices.Contains(allowedOrigins, origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Header("Access-Control-Max-Age", "86400")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// StripTrailingSlash normalises request paths so /tasks/ and /tasks both route correctly.
func StripTrailingSlash() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if len(path) > 1 && strings.HasSuffix(path, "/") {
			c.Request.URL.Path = path[:len(path)-1]
		}
		c.Next()
	}
}
